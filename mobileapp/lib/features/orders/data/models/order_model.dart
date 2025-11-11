import '../../../goods/data/models/good_model.dart';
import '../../domain/entities/order_entity.dart';

class OrderModel extends OrderEntity {
  const OrderModel({
    required super.id,
    required super.userId,
    required super.goodId,
    super.instanceId,
    required super.status,
    required super.createdAt,
    super.good,
  });

  factory OrderModel.fromJson(Map<String, dynamic> json) {
    try {
      return OrderModel(
        id: json['id'] as String,
        userId: json['user_id'] as String,
        goodId: json['good_id'] as String,
        instanceId: json['instance_id'] as String?,
        status: json['status'] as String,
        createdAt: json['created_at'] != null && json['created_at'] != ''
            ? DateTime.parse(json['created_at'] as String)
            : DateTime.now(),
        good: json['good'] != null
            ? GoodModel.fromJson(json['good'] as Map<String, dynamic>)
            : null,
      );
    } catch (e) {
      rethrow;
    }
  }

  Map<String, dynamic> toJson() {
    return {
      'id': id,
      'user_id': userId,
      'good_id': goodId,
      if (instanceId != null) 'instance_id': instanceId,
      'status': status,
      'created_at': createdAt.toIso8601String(),
      if (good != null) 'good': (good as GoodModel).toJson(),
    };
  }

  OrderEntity toEntity() {
    return OrderEntity(
      id: id,
      userId: userId,
      goodId: goodId,
      instanceId: instanceId,
      status: status,
      createdAt: createdAt,
      good: good,
    );
  }
}
