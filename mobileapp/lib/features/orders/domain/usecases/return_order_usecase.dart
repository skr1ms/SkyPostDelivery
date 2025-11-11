import '../repositories/orders_repository.dart';

class ReturnOrderUseCase {
  final OrdersRepository repository;

  const ReturnOrderUseCase(this.repository);

  Future<void> call(String orderId) async {
    await repository.returnOrder(orderId);
  }
}

