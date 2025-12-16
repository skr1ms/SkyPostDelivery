import '../../domain/entities/order_entity.dart';
import '../../domain/repositories/orders_repository.dart';
import '../datasources/orders_remote_datasource.dart';

class OrdersRepositoryImpl implements OrdersRepository {
  final OrdersRemoteDataSource remoteDataSource;

  const OrdersRepositoryImpl(this.remoteDataSource);

  @override
  Future<OrderEntity> createOrder({
    required String userId,
    required String goodId,
  }) async {
    final model = await remoteDataSource.createOrder(
      userId: userId,
      goodId: goodId,
    );
    return model.toEntity();
  }

  @override
  Future<List<OrderEntity>> createMultipleOrders({
    required String userId,
    required List<String> goodIds,
  }) async {
    final models = await remoteDataSource.createMultipleOrders(
      userId: userId,
      goodIds: goodIds,
    );
    return models.map((model) => model.toEntity()).toList();
  }

  @override
  Future<OrderEntity> getOrder(String orderId) async {
    final model = await remoteDataSource.getOrder(orderId);
    return model.toEntity();
  }

  @override
  Future<List<OrderEntity>> getUserOrders(String userId) async {
    final models = await remoteDataSource.getUserOrders(userId);
    return models.map((model) => model.toEntity()).toList();
  }

  @override
  Future<void> returnOrder(String orderId) async {
    await remoteDataSource.returnOrder(orderId);
  }
}

