import '../repositories/orders_repository.dart';
import '../entities/order_entity.dart';

class CreateMultipleOrdersUseCase {
  final OrdersRepository repository;

  CreateMultipleOrdersUseCase(this.repository);

  Future<List<OrderEntity>> call({
    required String userId,
    required List<String> goodIds,
  }) async {
    return await repository.createMultipleOrders(
      userId: userId,
      goodIds: goodIds,
    );
  }
}
