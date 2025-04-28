# gRPC Endpoint Query Examples

The following queries assume your gRPC server is running on `localhost:50051` and that the service name is `order.OrderService` with the following RPC methods defined in your proto:

- **ConfirmOrder(CreateOrderRequest) returns (CreateOrderResponse)**
- **GetOrderByID(GetOrderByIDRequest) returns (GetOrderByIDResponse)**
- **ListOrders(ListOrdersRequest) returns (ListOrdersResponse)**
- **ProcessOrder(ProcessOrderRequest) returns (ProcessOrderResponse)**
- **ReturnOrder(ReturnOrderRequest) returns (ReturnOrderResponse)**

> **Note:** For wrapper types (e.g. `google.protobuf.Int32Value`), pass the underlying value directly (like `limit: 10`).

---

## 1. Confirm Order

```bash
grpcurl -plaintext -d '{
  "order_id": 123,
  "user_id": 456,
  "expiration_time": "2027-03-10T15:00:00Z",
  "weight": 10,
  "cost": 200,
  "package_type": "box",
  "is_additional_film": true
}' localhost:50051 order.OrderService/ConfirmOrder
```

## 2. List Orders
```bash
grpcurl -plaintext -d '{
  "user_id": 456,
  "last_id": 0,
  "limit": 10
}' localhost:50051 order.OrderService/ListOrders
```

## 3. List All Orders
```bash
grpcurl -plaintext -d '{}' localhost:50051 order.OrderService/ListOrders
```

## 4. List All Orders With Pagination
```bash
grpcurl -plaintext -d '{
  "last_id": 10,
  "limit": 20
}' localhost:50051 order.OrderService/ListOrders
```

## 5. List Refunded Orders
```bash
grpcurl -plaintext -d '{
  "status": "refunded",
  "last_id": 0,
  "limit": 10
}' localhost:50051 order.OrderService/ListOrders
```

## 6. Get Order By Id
```bash
grpcurl -plaintext -d '{
  "order_id": 123
}' localhost:50051 order.OrderService/GetOrderByID
```

## 7. Return Order
```bash
grpcurl -plaintext -d '{
  "order_id": 123
}' localhost:50051 order.OrderService/ReturnOrder
```

## 8. Process Order (Complete)
```bash
grpcurl -plaintext -d '{
  "order_id": 123,
  "user_id": 456,
  "action": "complete"
}' localhost:50051 order.OrderService/ProcessOrder
```

## 9. Process Order (Refund)
```bash
grpcurl -plaintext -d '{
  "order_id": 123,
  "user_id": 456,
  "action": "refund"
}' localhost:50051 order.OrderService/ProcessOrder
```

## 10. Search Term
```bash
grpcurl -plaintext -d '{
  "search_term": "1"
}' localhost:50051 order.OrderService/ListOrders
```