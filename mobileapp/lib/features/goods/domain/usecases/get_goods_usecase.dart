import '../entities/good_entity.dart';
import '../repositories/goods_repository.dart';

class GetGoodsUseCase {
  final GoodsRepository repository;

  const GetGoodsUseCase(this.repository);

  Future<List<GoodEntity>> call() {
    return repository.getGoods();
  }
}

