import '../entities/order_entity.dart';
import '../repositories/orders_repository.dart';

class CreateOrderUseCase {
  final OrdersRepository repository;

  const CreateOrderUseCase(this.repository);

  Future<OrderEntity> call({
    required String userId,
    required String goodId,
  }) {
    return repository.createOrder(userId: userId, goodId: goodId);
  }
}

