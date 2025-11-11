import '../../domain/entities/good_entity.dart';
import '../../domain/repositories/goods_repository.dart';
import '../datasources/goods_remote_datasource.dart';

class GoodsRepositoryImpl implements GoodsRepository {
  final GoodsRemoteDataSource remoteDataSource;

  const GoodsRepositoryImpl(this.remoteDataSource);

  @override
  Future<List<GoodEntity>> getGoods() async {
    final models = await remoteDataSource.getGoods();
    return models.map((model) => model.toEntity()).toList();
  }
}

