import '../../../goods/domain/entities/good_entity.dart';

class OrderEntity {
  final String id;
  final String userId;
  final String goodId;
  final String? instanceId;
  final String status;
  final DateTime createdAt;
  final GoodEntity? good;

  const OrderEntity({
    required this.id,
    required this.userId,
    required this.goodId,
    this.instanceId,
    required this.status,
    required this.createdAt,
    this.good,
  });

  String get statusRu {
    switch (status.toLowerCase()) {
      case 'pending':
        return 'В пути';
      case 'delivered':
        return 'Прибыл';
      case 'completed':
        return 'Получен';
      case 'cancelled':
        return 'Отменен';
      default:
        return status;
    }
  }

  String get statusColor {
    switch (status.toLowerCase()) {
      case 'pending':
        return '#FFC107';
      case 'delivered':
        return '#4CAF50';
      case 'completed':
        return '#26A69A';
      case 'cancelled':
        return '#EF5350';
      default:
        return '#9E9E9E';
    }
  }
}
