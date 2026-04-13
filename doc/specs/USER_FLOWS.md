# User Flows Specification

This document maps the abstract platform user flows defined in `user_flows.txt` onto explicit backend database mutation sequences executed by the inner `/pkg/services/` tier.

## REQ-CHECKOUT-01: Multi-Producer Checkout to Order Confirmation (UF-CUST-2.4)
When a customer clicks 'Checkout', the platform executes a complex multi-stage pipeline supporting products from multiple producers in a single transaction:

### Phase 1: `ProcessCheckout` (State: Pending)
1. **Group by Producer**: Cart items are grouped by `Product.UserID` (producer)
2. **DB BATCH INSERT**: For each producer group:
   - Creates a new `Order` record with `Status`="pending" and `PaymentStatus`="pending"
   - Assigns a shared `OrderGroupID` to all orders in this checkout
   - Generates immutable `OrderItem` records tied strictly to the new Order ID
   - Generates a 1:1 `Payment` record with `Status`="pending"
   - Creates a 1:1 `Shipping` record with shipping details from checkout
3. **OrderGroupSummary**: Returns all orders grouped by producer with individual totals

### Phase 2: `ConfirmOrderGroup` (State: Locked & Paid)
4. **DB UPDATE**: For each order in the group:
   - Iterates each `OrderItem` and executes decrements against the active `Product.Stock`
   - Appends the internal `TransactionID` to the `Payment` record and upgrades status to `"paid"`
   - Migrates the parent `Order` mapping setting `Status`="processing" and `PaymentStatus`="paid"
5. **DB DELETE**: Wipes the mutable session state by completely clearing the User's `Cart`

---

## REQ-CHECKOUT-02: Partial Producer Decline (Multi-Producer Scenario)
When one producer declines/cancels their order while others confirm:

1. **Producer Cancellation**: Producer calls `/orders/cancel` with their `order_id`
2. **DB UPDATE**: 
   - Sets `Order.Status` = "canceled"
   - Sets `Order.PaymentStatus` = "refunded"
   - Updates `Payment.Status` = "refunded"
3. **Refund Processing**: The specific order's payment is refunded (not the entire group)
4. **Other Orders Unaffected**: Orders from other producers remain `processing` status

**Test Coverage**: `TestMultiProducerOrderWithOneDecline` validates this scenario.

---

## REQ-CHECKOUT-03: Full Producer Decline (All Orders Canceled)
When all producers cancel their orders:

1. **Sequential Cancellations**: Each producer cancels their respective order via `/orders/cancel`
2. **Full Refund**: Each order's payment is refunded independently
3. **Customer View**: Customer sees all orders as "canceled" with "refunded" payment status

**Test Coverage**: `TestMultiProducerOrderAllDeclined` validates this scenario.

---

## REQ-REVIEW-01: Verified Purchase Check (UF-CUST-2.6)
To secure the marketplace, the `ReviewService.AddReview` runs a defensive logic scan ensuring only verifiable customers may review products:

1. **Order Scan**: Extracts an array of every `Order` inherently tied to the active `UserID` making the request.
2. **Details Expansion**: Executes rolling queries across the array expanding each node via `GetOrderWithDetails`.
3. **Product Hash Hunt**: Loops through every underlying `OrderItem` map structurally comparing the itemized `ProductID` with the requested `ReviewDTO.ProductID`.
4. **Blockade**: If the nested loop completes without flipping the `purchased` boolean to true, the request throws an isolated error: `"user has not purchased this product"` and restricts the DB write. If the boolean intercepts true, it naturally flows into a review DB insertion snippet. 

---

## REQ-FULFILL-01: Product Fulfillment & Tracking (UF-PROD-3.3)
When a Producer signals they are sending the package through the `OrderService.FulfillOrder` pipeline, the following data operations occur:

1. **Ownership Constraint Check**: Scans `OrderItem` mapping details down against their underlying `Product` entities confirming at least one product `UserID` identically matches the authenticated `ProducerID`. 
2. **Financial Validation**: Asserts the structural integrity is secure (`Order.PaymentStatus` == "paid"). An unpaid order strictly prohibits fulfillment tracking overrides.
3. **DB UPDATE**: Updates the parent `Order.Status` with a verified schema whitelist switch (falling safely onto `"shipped"`).
4. **DB UPDATE**: Because the state triggered a `"shipped"` transition, a secondary branch halts execution unless a strict `TrackingCode` payload is present. It injects a runtime `time.Now()` into the `ShippedAt` attribute then runs `UpdateShippingTracking` to persist the shipping metadata seamlessly onto the `Shipping` database record.
