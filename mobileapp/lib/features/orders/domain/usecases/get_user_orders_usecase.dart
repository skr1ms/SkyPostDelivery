import '../entities/order_entity.dart';
import '../repositories/orders_repository.dart';

class GetUserOrdersUseCase {
  final OrdersRepository repository;

  const GetUserOrdersUseCase(this.repository);

  Future<List<OrderEntity>> call(String userId) {
    return repository.getUserOrders(userId);
  }
}

