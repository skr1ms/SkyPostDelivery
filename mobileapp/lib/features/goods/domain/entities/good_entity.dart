class GoodEntity {
  final String id;
  final String name;
  final double weight;
  final double height;
  final double length;
  final double width;
  final int quantityAvailable;

  const GoodEntity({
    required this.id,
    required this.name,
    required this.weight,
    required this.height,
    required this.length,
    required this.width,
    required this.quantityAvailable,
  });

  double get volume => height * length * width;

  String get dimensions =>
      '${length.toInt()}x${width.toInt()}x${height.toInt()} см';

  bool get isAvailable => quantityAvailable > 0;
}
