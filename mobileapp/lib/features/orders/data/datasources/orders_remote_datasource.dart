import '../../../../core/network/http_client.dart';
import '../../../../core/network/api_constants.dart';
import '../models/order_model.dart';

abstract class OrdersRemoteDataSource {
  Future<OrderModel> createOrder({
    required String userId,
    required String goodId,
  });

  Future<List<OrderModel>> createMultipleOrders({
    required String userId,
    required List<String> goodIds,
  });

  Future<OrderModel> getOrder(String orderId);

  Future<List<OrderModel>> getUserOrders(String userId);

  Future<void> returnOrder(String orderId);
}

class OrdersRemoteDataSourceImpl implements OrdersRemoteDataSource {
  final HttpClient httpClient;

  const OrdersRemoteDataSourceImpl(this.httpClient);

  @override
  Future<OrderModel> createOrder({
    required String userId,
    required String goodId,
  }) async {
    final response = await httpClient.post(
      ApiConstants.orders,
      body: {'good_id': goodId},
      requiresAuth: true,
    );
    return OrderModel.fromJson(response);
  }

  @override
  Future<List<OrderModel>> createMultipleOrders({
    required String userId,
    required List<String> goodIds,
  }) async {
    final response = await httpClient.postList(
      ApiConstants.ordersBatch,
      body: {'good_ids': goodIds},
      requiresAuth: true,
    );

    return response
        .map((json) => OrderModel.fromJson(json as Map<String, dynamic>))
        .toList();
  }

  @override
  Future<OrderModel> getOrder(String orderId) async {
    final response = await httpClient.get(
      ApiConstants.orderById(orderId),
      requiresAuth: true,
    );
    return OrderModel.fromJson(response);
  }

  @override
  Future<List<OrderModel>> getUserOrders(String userId) async {
    final response = await httpClient.getList(
      ApiConstants.ordersByUser(userId),
      requiresAuth: true,
    );


    try {
      return response.map((json) {
        return OrderModel.fromJson(json as Map<String, dynamic>);
      }).toList();
    } catch (e) {
      rethrow;
    }
  }

  @override
  Future<void> returnOrder(String orderId) async {
    await httpClient.post(
      ApiConstants.returnOrder(orderId),
      body: {},
      requiresAuth: true,
    );
  }
}
