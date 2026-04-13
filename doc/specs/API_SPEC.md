# API Specifications

This specification maps the current application endpoints divided by functional domains, establishing the DTOs used for I/O and identifying authorization needs. It acts as the 'Current Truth', noting where actual Go handler implementations dynamically differ from `openapi.json`.

## Universal Security Rules
- **JWT (Authentication)**: Endpoints marked with **[Auth]** require a valid `jwt` HttpOnly cookie. The context extracts the user session using `middleware.GetUserFromContext`.
- **CSRF (State Mutations)**: Endpoints marked with **[CSRF]** (typically `POST`, `PUT`, `DELETE` operations) require the `X-CSRF-Token` header.

---

## 1. User Management (`pkg/handlers/userHandler.go`)
| Method | Endpoint | Security | DTO Request Payload | Description |
|:---|:---|:---|:---|:---|
| **POST** | `/users/register`       | [CSRF]         | `models.UserDTO` | Registers a new user. |
| **POST** | `/users/login`          | [CSRF]         | `{ email, password }` | Authenticates a user and sets the JWT. |
| **GET**  | `/users/get-by-id`      | [Auth]         | *Query parameters* | Retrieves a user by their ID. |
| **GET**  | `/users/get-by-name`    | [Auth]         | *Query parameters* | Retrieves a user by their name. |
| **GET**  | `/users/get-by-email`   | [Auth]         | *Query parameters* | Retrieves a user by their email. |
| **PUT**  | `/users/update`         | [Auth], [CSRF] | `models.UserDTO` | Updates the authenticated user's profile. |
| **DELETE**| `/users/delete`        | [Auth], [CSRF] | *None* | Wipes the active user from the system. |

---

## 2. Auth & Session Management (`pkg/handlers/authHandler.go`)
| Method | Endpoint | Security | DTO Request Payload | Description |
|:---|:---|:---|:---|:---|
| **POST** | `/auth/change-password` | [Auth], [CSRF] | `models.ChangePasswordDTO` | Updates an authenticated user's password. |
| **POST** | `/auth/logout`          | [Auth], [CSRF] | *None* | Flushes the `jwt` and `csrf_token` cookies. |

---

## 3. Product Management (`pkg/handlers/productHandler.go`)
| Method | Endpoint | Security | DTO Request Payload | Description |
|:---|:---|:---|:---|:---|
| **POST** | `/products/create`      | [Auth], [CSRF] | `models.ProductDTO` | Creates a new product. User takes ownership. |
| **GET**  | `/product?id=...`       | *None*         | *Query parameters* | Fetch a specific product by ID. |
| **GET**  | `/products`             | *None*         | *Query parameters* | Fetch a broad catalog with optional filters. |
| **GET**  | `/products/category?...`| *None*         | *Query parameters* | Categorical product query. |
| **GET**  | `/products/my-products` | [Auth]         | *None* | Fetches products matching `authUser.ID`. |
| **PUT**  | `/products/update?id=..`| [Auth], [CSRF] | `models.ProductDTO` | Modifies product fields (Ownership validated). |
| **DELETE**| `/products/delete?id=..`| [Auth], [CSRF] | *None* | Restricts deletion to the owning producer. |

---

## 4. Cart Management (`pkg/handlers/cartHandler.go`)

> [!NOTE]
> **Optional Auth**: Cart routes utilize `OptionalAuthenticateJWT`. If the user is logged in, their persistent DB cart is targeted. If unauthenticated, a temporary session cookie generates a guest cart.

| Method | Endpoint | Security | DTO Request Payload | Description |
|:---|:---|:---|:---|:---|
| **POST** | `/cart/add`             | [Optional Auth], [CSRF] | `[]models.CartItemDTO` | Adds a batch of items to the cart. |
| **POST** | `/cart/remove`          | [Optional Auth], [CSRF] | `[]models.CartItemDTO` | Removes targeted items from the cart. |
| **POST** | `/cart/clear`           | [Optional Auth], [CSRF] | *None* | Completely clears all items from the cart. |
| **GET**  | `/cart/get`             | [Optional Auth]         | *None* | Retrieves all current valid cart items. |
| **POST** | `/cart/update`          | [Optional Auth], [CSRF] | `[]models.CartItemDTO` | Updates quantities for existing cart items. |

---

## 5. Order Management & Checkout (`pkg/handlers/orderHandler.go`)

> [!WARNING]
> ### 🛑 DISCREPANCY DETECTED: `GetSalesReport`
> **OpenAPI states**: `/orders/sales-report` is a `GET` request.
> **Current Go Truth**: `OrderHandler.GetSalesReport` processes incoming data by reading the request payload `json.NewDecoder(r.Body).Decode(&request)` into a `models.SalesReportRequest`.
> *Issue*: Fetching a JSON body across a `GET` HTTP request is an anti-pattern often dropped by networking layers. The Go code dictates behavior, so this reflects the Current Truth, but this endpoint should likely be migrated to `POST` or mapped to Query Parameters.

| Method | Endpoint | Security | DTO Request Payload | Description |
|:---|:---|:---|:---|:---|
| **POST** | `/orders/checkout`      | [Auth], [CSRF] | `models.CheckoutRequest` | Converts a Cart into an OrderGroupSummary (orders grouped by producer). |
| **POST** | `/orders/confirm`       | [Auth], [CSRF] | `{ OrderGroupID: uuid.UUID }` | Finalizes all orders in the group, creating Order, Payment, and Shipping records. |
| **POST** | `/orders/group-details` | [Auth]         | `{ OrderGroupID: uuid.UUID }` | Retrieve deep order details for all orders in a group (multi-producer). |
| **POST** | `/orders/summary`       | [Auth]         | `{ OrderID: uuid.UUID }` | Submits validation requests to review an order summary. |
| **POST** | `/orders/details`       | [Auth]         | `{ OrderID: uuid.UUID }` | Retrieve deep order details for Customer audits. |
| **GET**  | `/orders/user`          | [Auth]         | *None* | Retrieves all previous orders for current customer. |
| **GET**  | `/orders/producer`      | [Auth]         | *None* | Filtered metrics requiring `user.IsProducer` flag to be true. |
| **POST** | `/orders/fulfill`       | [Auth], [CSRF] | `models.OrderFulfillmentRequest` | Producers update shipping statuses statically. |
| **GET**  | `/orders/sales-report`  | [Auth]         | `models.SalesReportRequest` *(See Discrepancy)* | Fetches revenue matrices based on data filters. |
| **POST** | `/orders/cancel`        | [Auth], [CSRF] | `{ OrderID: uuid.UUID }` | Allows producers to formally cancel orders (triggers refund). |
| **POST** | `/orders/delete`        | [Auth], [CSRF] | `{ OrderID: uuid.UUID }` | Soft delete functionality for customers interacting with their log. |

---

> [!NOTE]
> ### Multi-Producer Order Flow
> - `/orders/checkout` creates one order per producer from the cart items
> - All orders are linked via a common `order_group_id`
> - `/orders/confirm` takes `order_group_id` and confirms all orders at once
> - Each producer can cancel only their own order (triggers refund for that order)
> - Customer can view all orders individually via `/orders/details`
> - Full group details via `/orders/group-details`

---

## 6. Review Interactions (`pkg/handlers/reviewHandler.go`)
| Method | Endpoint | Security | DTO Request Payload | Description |
|:---|:---|:---|:---|:---|
| **POST** | `/reviews/add`          | [Auth], [CSRF] | `models.ReviewDTO` | Casts a 1-5 rating constraint against a product. |
| **GET**  | `/reviews/get?id=...`   | *None*         | *Query parameters* | Retrieves global review matrices. |
| **PUT**  | `/reviews/update?id=..` | [Auth], [CSRF] | `{ Rating: int, Comment: *string }` | Overwrites a review dynamically. |
| **DELETE**| `/reviews/delete?id=..`| [Auth], [CSRF] | *None* | Wipes a review statically from the catalog map. |

---

## 7. Company Management (`pkg/handlers/companyHandler.go`)
| Method | Endpoint | Security | DTO Request Payload | Description |
|:---|:---|:---|:---|:---|
| **POST** | `/companies/create`      | [Auth], [CSRF] | `models.CompanyRequest` | Creates a new company profile linked to the producer. |
| **GET**  | `/companies/get-by-user` | [Auth]         | *None* | Retrieves the company profile for the authenticated producer. |
| **GET**  | `/companies/get-by-id`   | *None*         | *Query parameters* | Fetch a specific company by its ID publicly. |
| **PUT**  | `/companies/update`      | [Auth], [CSRF] | `models.CompanyRequest` | Updates the authenticated producer's company record. |
| **DELETE**| `/companies/delete`     | [Auth], [CSRF] | *None* | Deletes the company record. |
