# Database Model Specifications

*Note: This project relies natively on standard SQL mapping via the core `database/sql` driver alongside auxiliary libraries like `sqlx`. All database interactions strictly expect standard SQL parameterization. New application models MUST include native `db:"..."` tags rather than GORM implementations to synergize with the current Repository layer architecture.*

## The Checkout Chain

The primary transactional block stems from the `Order`. When an order is completed, it rigidly links external data:
- **Order to Items (`1:N`)**: `Order.ID` -> `OrderItem.OrderID`. Maps individual product snapshots to the transaction.
- **Order to Payment (`1:1`)**: `Order.ID` -> `Payment.OrderID`. Handles transaction states dynamically.
- **Order to Shipping (`1:1`)**: `Order.ID` -> `Shipping.OrderID`. Directs logistical tracking code events per transaction.

### Multi-Producer Checkout Flow

When a customer checkout includes products from multiple producers:
1. One `Order` is created **per producer** (not per order group)
2. All orders share the same `OrderGroupID` to link them logically
3. Each order has its own `Payment` and `Shipping` records
4. Each producer can independently accept, fulfill, or cancel their specific order
5. If a producer cancels, only that order is refunded (not the entire group)
6. The customer sees all orders individually in their order history

---

### User
| Field        | Go Type      | Relations / FK | Database Mapping (db tags) |
|--------------|--------------|----------------|----------------------------|
| `ID`         | `uuid.UUID`  |                | `id`                       |
| `FirstName`  | `string`     |                | `first_name`               |
| `LastName`   | `string`     |                | `last_name`                |
| `MiddleName` | `string`     |                | `middle_name`              |
| `Email`      | `string`     |                | `email`                    |
| `Phone`      | `int64`      |                | `phone`                    |
| `IsProducer` | `bool`       |                | `is_producer`              |
| `Address`    | `string`     |                | `address`                  |
| `City`       | `string`     |                | `city`                     |
| `Country`    | `string`     |                | `country`                  |
| `ZipCode`    | `int32`      |                | `zip_code`                 |
| `CreatedAt`  | `time.Time`  |                | `created_at`               |
| `UpdatedAt`  | `*time.Time` |                | `updated_at`               |

### Auth
| Field        | Go Type      | Relations / FK           | Database Mapping (db tags) |
|--------------|--------------|--------------------------|----------------------------|
| `ID`         | `uuid.UUID`  |                          | `id`                       |
| `UserID`     | `uuid.UUID`  | FK -> `User.ID`          | `user_id`                  |
| `CreatedAt`  | `time.Time`  |                          | `created_at`               |
| `Active`     | `bool`       |                          | `active`                   |
| `Password`   | `string`     |                          | `password`                 |
| `UpdatedAt`  | `*time.Time` |                          | `updated_at`               |

### Product
| Field         | Go Type      | Relations / FK             | Database Mapping (db tags) |
|---------------|--------------|----------------------------|----------------------------|
| `ID`          | `uuid.UUID`  |                            | `id`                       |
| `Name`        | `string`     |                            | `name`                     |
| `Description` | `string`     |                            | `description`              |
| `Price`       | `float64`    |                            | `price`                    |
| `Stock`       | `int32`      |                            | `stock`                    |
| `Category`    | `string`     |                            | `category`                 |
| `ImageUrl`    | `string`     |                            | `image_url`                |
| `CreatedAt`   | `time.Time`  |                            | `created_at`               |
| `UpdatedAt`   | `*time.Time` |                            | `updated_at`               |
| `UserID`      | `uuid.UUID`  | FK -> `User.ID` (Producer) | `user_id`                  |

### Company
| Field           | Go Type      | Relations / FK             | Database Mapping (db tags) |
|-----------------|--------------|----------------------------|----------------------------|
| `ID`            | `uuid.UUID`  |                            | `id`                       |
| `UserID`        | `uuid.UUID`  | FK -> `User.ID` (Producer) | `user_id`                  |
| `Name`          | `string`     |                            | `name`                     |
| `Address`       | `*string`    |                            | `address`                  |
| `City`          | `*string`    |                            | `city`                     |
| `Country`       | `*string`    |                            | `country`                  |
| `ZipCode`       | `*string`    |                            | `zip_code`                 |
| `ReviewAverage` | `float64`    |                            | `review_average`           |
| `ReviewCount`   | `int`        |                            | `review_count`             |
| `CreatedAt`     | `time.Time`  |                            | `created_at`               |
| `UpdatedAt`     | `*time.Time` |                            | `updated_at`               |

### Order
| Field           | Go Type      | Relations / FK           | Database Mapping (db tags) |
|-----------------|--------------|--------------------------|----------------------------|
| `ID`            | `uuid.UUID`  |                          | `id`                       |
| `OrderGroupID`  | `*uuid.UUID` | FK -> `OrderGroup.ID`    | `order_group_id`           |
| `UserID`        | `uuid.UUID`  | FK -> `User.ID`          | `user_id`                  |
| `TotalAmount`   | `float64`    |                          | `total_amount`             |
| `Status`        | `string`     |                          | `status`                   |
| `PaymentStatus` | `string`     |                          | `payment_status`           |
| `CreatedAt`     | `time.Time`  |                          | `created_at`               |
| `UpdatedAt`     | `*time.Time` |                          | `updated_at`               |

> **Note**: Multiple orders from different producers can be grouped via `OrderGroupID` to enable multi-producer checkout flow.

### OrderItem
| Field       | Go Type     | Relations / FK           | Database Mapping (db tags) |
|-------------|-------------|--------------------------|----------------------------|
| `ID`        | `uuid.UUID` |                          | `id`                       |
| `OrderID`   | `uuid.UUID` | FK -> `Order.ID`         | `order_id`                 |
| `ProductID` | `uuid.UUID` | FK -> `Product.ID`       | `product_id`               |
| `Quantity`  | `int`       |                          | `quantity`                 |
| `Price`     | `float64`   |                          | `price`                    |

### Payment
| Field           | Go Type      | Relations / FK           | Database Mapping (db tags) |
|-----------------|--------------|--------------------------|----------------------------|
| `ID`            | `uuid.UUID`  |                          | `id`                       |
| `OrderID`       | `uuid.UUID`  | FK -> `Order.ID`         | `order_id`                 |
| `Amount`        | `float64`    |                          | `amount`                   |
| `PaymentMethod` | `string`     |                          | `payment_method`           |
| `Status`        | `string`     |                          | `status`                   |
| `TransactionID` | `*string`    |                          | `transaction_id`           |
| `CreatedAt`     | `time.Time`  |                          | `created_at`               |

### Shipping
| Field          | Go Type      | Relations / FK           | Database Mapping (db tags) |
|----------------|--------------|--------------------------|----------------------------|
| `ID`           | `uuid.UUID`  |                          | `id`                       |
| `OrderID`      | `uuid.UUID`  | FK -> `Order.ID`         | `order_id`                 |
| `Method`       | `*string`    |                          | `method`                   |
| `TrackingCode` | `*string`    |                          | `tracking_code`            |
| `Cost`         | `*float64`   |                          | `cost`                     |
| `Address`      | `string`     |                          | `address`                  |
| `City`         | `string`     |                          | `city`                     |
| `Country`      | `string`     |                          | `country`                  |
| `ZipCode`      | `string`     |                          | `zip_code`                 |
| `ShippedAt`    | `*time.Time` |                          | `shipped_at`               |
| `DeliveredAt`  | `*time.Time` |                          | `delivered_at`             |
| `CreatedAt`    | `time.Time`  |                          | `created_at`               |
| `UpdatedAt`    | `*time.Time` |                          | `updated_at`               |

### Cart
| Field       | Go Type      | Relations / FK           | Database Mapping (db tags) |
|-------------|--------------|--------------------------|----------------------------|
| `ID`        | `uuid.UUID`  |                          | `id`                       |
| `UserID`    | `*uuid.UUID` | FK -> `User.ID`          | `user_id`                  |
| `CreatedAt` | `time.Time`  |                          | `created_at`               |
| `UpdatedAt` | `*time.Time` |                          | `updated_at`               |
| `SessionID` | `*string`    | Guest carts              | `session_id`               |

### CartItem
| Field       | Go Type     | Relations / FK           | Database Mapping (db tags) |
|-------------|-------------|--------------------------|----------------------------|
| `ID`        | `uuid.UUID` |                          | `id`                       |
| `CartID`    | `uuid.UUID` | FK -> `Cart.ID`          | `cart_id`                  |
| `ProductID` | `uuid.UUID` | FK -> `Product.ID`       | `product_id`               |
| `Quantity`  | `int`       |                          | `quantity`                 |
| `Price`     | `float64`   |                          | `price`                    |
| `CreatedAt` | `time.Time` |                          | `created_at`               |

### Review
| Field       | Go Type      | Relations / FK           | Database Mapping (db tags) |
|-------------|--------------|--------------------------|----------------------------|
| `ID`        | `uuid.UUID`  |                          | `id`                       |
| `ProductID` | `uuid.UUID`  | FK -> `Product.ID`       | `product_id`               |
| `UserID`    | `uuid.UUID`  | FK -> `User.ID`          | `user_id`                  |
| `Rating`    | `int`        |                          | `rating`                   |
| `Comment`   | `*string`    |                          | `comment`                  |
| `CreatedAt` | `time.Time`  |                          | `created_at`               |
