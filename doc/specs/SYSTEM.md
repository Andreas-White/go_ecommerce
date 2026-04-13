# System Architecture & Business Logic

## Dual-Role System

The platform operates using a role-based partition determined by the `IsProducer` flag on the `User` accounts:

- **Customer (`IsProducer: false`)**: The default consumer persona. Customers are permitted to browse products, manage shopping carts (authenticated or guest sessions), post product reviews, and execute checkouts.
- **Producer (`IsProducer: true`)**: The vendor persona. A producer has strict ownership over the products they list. Producers have authorization to manage their product catalog (prices, stock, metadata), update shipping/fulfillment statuses for incoming orders, and access detailed, role-restricted sales reports.

*Actionable Note for Future Development*: Enhancements involving `user_name` customizations for all users and extended `company_details` integrations for producers are planned workflow upgrades. When implemented, they MUST strictly abide by the pattern established in the existing `userService`, traversing the exact Handler -> Service -> Repository tiers.

## Multi-Producer Order Architecture

The system supports checkout scenarios where a customer purchases products from multiple producers in a single transaction:

### Order Grouping Logic
- **One Order Per Producer**: When a customer checks out with items from N different producers, N separate `Order` records are created (one per producer)
- **Shared OrderGroupID**: All orders in a single checkout share the same `OrderGroupID` to maintain logical grouping
- **Independent Processing**: Each order maintains its own:
  - `Payment` record (can be independently refunded)
  - `Shipping` record (independent tracking)
  - Status workflow (accept, fulfill, cancel)

### Producer Order Independence
- Producers can only view and manage orders containing their products
- Producers can independently cancel their orders without affecting other producers' orders
- Cancellation triggers automatic refund for that specific order only
- Other orders in the group continue processing normally

### Order Status Flow
| Status | Description |
|--------|-------------|
| `pending` | Order created, awaiting customer confirmation |
| `processing` | Confirmed, payment successful, awaiting fulfillment |
| `accepted` | Producer accepted the order |
| `shipped` | Producer shipped with tracking |
| `canceled` | Producer canceled (triggers refund) |

### Payment Status Flow
| Status | Description |
|--------|-------------|
| `pending` | Awaiting payment confirmation |
| `paid` | Payment successful |
| `refunded` | Payment refunded (via producer cancellation) |

## Immutable Checkout Pipeline

The order generation workflow enforces a strict immutability chain, locking down state at the point of translation from Cart to Order:
1. **Dynamic Cart Phase**: Items (`CartItem`) are dynamically modified, removed, or updated in `Cart` relative to active stock limits.
2. **Checkout Execution**: When checkout triggers, an `Order` parent entity is generated alongside immutable snapshots of products at that exact moment (stored as `OrderItem`). Price history changes no longer reflect on finalised `OrderItem`s.
3. **Record Anchoring**: Distinct `Payment` and `Shipping` records are immediately appended to the unique `OrderID`. 
4. **Data Sanity**: The `Order` cannot be directly reverted or edited post-payment initialization to uphold accurate financial audits. All alterations must occur via separate refund or order cancellation mechanisms rather than overwriting historical `Order` rows.
