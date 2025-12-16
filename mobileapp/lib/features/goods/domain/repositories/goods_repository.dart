import '../entities/good_entity.dart';

abstract class GoodsRepository {
  Future<List<GoodEntity>> getGoods();
}

