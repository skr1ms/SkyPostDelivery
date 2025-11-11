import '../../../../core/network/http_client.dart';
import '../../../../core/network/api_constants.dart';
import '../models/good_model.dart';

abstract class GoodsRemoteDataSource {
  Future<List<GoodModel>> getGoods();
}

class GoodsRemoteDataSourceImpl implements GoodsRemoteDataSource {
  final HttpClient httpClient;

  const GoodsRemoteDataSourceImpl(this.httpClient);

  @override
  Future<List<GoodModel>> getGoods() async {
    final response = await httpClient.getList(
      ApiConstants.goods,
      requiresAuth: true,
    );
    
    return response
        .map((json) => GoodModel.fromJson(json as Map<String, dynamic>))
        .toList();
  }
}

