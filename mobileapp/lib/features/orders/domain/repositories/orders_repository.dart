import '../entities/order_entity.dart';

abstract class OrdersRepository {
  Future<OrderEntity> createOrder({
    required String userId,
    required String goodId,
  });

  Future<List<OrderEntity>> createMultipleOrders({
    required String userId,
    required List<String> goodIds,
  });

  Future<OrderEntity> getOrder(String orderId);

  Future<List<OrderEntity>> getUserOrders(String userId);

  Future<void> returnOrder(String orderId);
}

