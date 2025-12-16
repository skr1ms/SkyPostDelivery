import '../../domain/entities/good_entity.dart';

class GoodModel extends GoodEntity {
  const GoodModel({
    required super.id,
    required super.name,
    required super.weight,
    required super.height,
    required super.length,
    required super.width,
    required super.quantityAvailable,
  });

  factory GoodModel.fromJson(Map<String, dynamic> json) {
    return GoodModel(
      id: json['id'] as String,
      name: json['name'] as String,
      weight: (json['weight'] as num).toDouble(),
      height: (json['height'] as num).toDouble(),
      length: (json['length'] as num).toDouble(),
      width: (json['width'] as num).toDouble(),
      quantityAvailable: json['quantity_available'] as int? ?? 0,
    );
  }

  Map<String, dynamic> toJson() {
    return {
      'id': id,
      'name': name,
      'weight': weight,
      'height': height,
      'length': length,
      'width': width,
      'quantity_available': quantityAvailable,
    };
  }

  GoodEntity toEntity() {
    return GoodEntity(
      id: id,
      name: name,
      weight: weight,
      height: height,
      length: length,
      width: width,
      quantityAvailable: quantityAvailable,
    );
  }
}
