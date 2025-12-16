import 'dart:async';
import 'package:flutter/material.dart';
import 'package:provider/provider.dart';
import 'package:font_awesome_flutter/font_awesome_flutter.dart';
import '../../../../core/theme/app_theme.dart';
import '../../../../core/di/injection_container.dart';
import '../../../auth/presentation/providers/auth_provider.dart';
import '../../domain/entities/order_entity.dart';
import '../../../auth/presentation/widgets/animated_background.dart';
import '../../../auth/presentation/widgets/glassmorphic_card.dart';

class OrdersScreen extends StatefulWidget {
  const OrdersScreen({super.key});

  @override
  State<OrdersScreen> createState() => _OrdersScreenState();
}

class _OrdersScreenState extends State<OrdersScreen>
    with WidgetsBindingObserver, AutomaticKeepAliveClientMixin {
  List<OrderEntity> _orders = [];
  bool _isLoading = true;
  String? _error;
  Timer? _refreshTimer;

  final _di = InjectionContainer();

  @override
  bool get wantKeepAlive => true;

  @override
  void initState() {
    super.initState();
    WidgetsBinding.instance.addObserver(this);
    _loadOrders();
    _startAutoRefresh();
  }

  @override
  void dispose() {
    WidgetsBinding.instance.removeObserver(this);
    _refreshTimer?.cancel();
    super.dispose();
  }

  @override
  void didChangeAppLifecycleState(AppLifecycleState state) {
    super.didChangeAppLifecycleState(state);
    if (state == AppLifecycleState.resumed && mounted) {
      _loadOrders(silent: true);
    }
  }

  void _startAutoRefresh() {
    _refreshTimer = Timer.periodic(const Duration(seconds: 10), (timer) {
      if (mounted) {
        _loadOrders(silent: true);
      }
    });
  }

  Future<void> _loadOrders({bool silent = false}) async {
    if (!silent) {
      setState(() {
        _isLoading = true;
        _error = null;
      });
    }

    try {
      final authProvider = Provider.of<AuthProvider>(context, listen: false);
      if (authProvider.user == null) {
        return;
      }

      final orders = await _di.getUserOrdersUseCase(authProvider.user!.id);

      if (mounted) {
        setState(() {
          _orders = orders;
          _error = null;
          if (!silent) {
            _isLoading = false;
          }
        });
      }
    } catch (e) {
      if (mounted && !silent) {
        setState(() {
          _error = e.toString();
          _isLoading = false;
        });
      }
    }
  }

  List<Map<String, dynamic>> _groupOrders(List<OrderEntity> orders) {
    final Map<String, Map<String, dynamic>> grouped = {};

    try {
      for (final order in orders) {
        final key =
            '${order.goodId}_${order.status}_${order.createdAt.year}_${order.createdAt.month}_${order.createdAt.day}_${order.createdAt.hour}_${order.createdAt.minute}';

        if (grouped.containsKey(key)) {
          grouped[key]!['count'] = (grouped[key]!['count'] as int) + 1;
          grouped[key]!['orders'].add(order);
        } else {
          grouped[key] = {
            'order': order,
            'count': 1,
            'orders': [order],
          };
        }
      }

      final result = grouped.values.toList()
        ..sort(
          (a, b) => (b['order'] as OrderEntity).createdAt.compareTo(
            (a['order'] as OrderEntity).createdAt,
          ),
        );
      return result;
    } catch (e) {
      rethrow;
    }
  }

  Color _getStatusColor(String status) {
    switch (status.toLowerCase()) {
      case 'pending':
        return const Color(0xFFFFC107);
      case 'delivered':
        return const Color(0xFF4CAF50);
      case 'completed':
        return const Color(0xFF26A69A);
      case 'cancelled':
        return const Color(0xFFEF5350);
      default:
        return const Color(0xFF9E9E9E);
    }
  }

  IconData _getStatusIcon(String status) {
    switch (status.toLowerCase()) {
      case 'pending':
        return FontAwesomeIcons.helicopter;
      case 'delivered':
        return FontAwesomeIcons.boxOpen;
      case 'completed':
        return FontAwesomeIcons.circleCheck;
      case 'cancelled':
        return FontAwesomeIcons.circleXmark;
      default:
        return FontAwesomeIcons.question;
    }
  }

  @override
  Widget build(BuildContext context) {
    super.build(context);
    return Scaffold(
      body: AnimatedBackground(
        child: SafeArea(
          child: Column(
            children: [
              // Header
              Padding(
                padding: const EdgeInsets.all(20),
                child: Row(
                  children: [
                    const Icon(
                      FontAwesomeIcons.clipboardList,
                      color: AppTheme.accentColor,
                      size: 28,
                    ),
                    const SizedBox(width: 12),
                    Text(
                      'Мои заказы',
                      style: Theme.of(context).textTheme.headlineMedium,
                    ),
                  ],
                ),
              ),

              // Content
              Expanded(
                child: _isLoading
                    ? const Center(child: CircularProgressIndicator())
                    : _error != null
                    ? Center(
                        child: Column(
                          mainAxisAlignment: MainAxisAlignment.center,
                          children: [
                            const Icon(
                              FontAwesomeIcons.triangleExclamation,
                              color: AppTheme.errorColor,
                              size: 48,
                            ),
                            const SizedBox(height: 16),
                            Text(
                              'Ошибка загрузки',
                              style: Theme.of(context).textTheme.titleLarge,
                            ),
                            const SizedBox(height: 24),
                            ElevatedButton.icon(
                              onPressed: _loadOrders,
                              icon: const Icon(FontAwesomeIcons.rotate),
                              label: const Text('Повторить'),
                            ),
                          ],
                        ),
                      )
                    : _orders.isEmpty
                    ? Center(
                        child: Column(
                          mainAxisAlignment: MainAxisAlignment.center,
                          children: [
                            Icon(
                              FontAwesomeIcons.boxOpen,
                              size: 64,
                              color: AppTheme.textSecondary.withValues(
                                alpha: 0.5,
                              ),
                            ),
                            const SizedBox(height: 16),
                            Text(
                              'Нет заказов',
                              style: Theme.of(context).textTheme.titleLarge
                                  ?.copyWith(color: AppTheme.textSecondary),
                            ),
                            const SizedBox(height: 8),
                            Text(
                              'Сделайте первый заказ в каталоге',
                              style: Theme.of(context).textTheme.bodyMedium
                                  ?.copyWith(color: AppTheme.textSecondary),
                            ),
                          ],
                        ),
                      )
                    : RefreshIndicator(
                        onRefresh: _loadOrders,
                        child: Builder(
                          builder: (context) {
                            final groupedOrders = _groupOrders(_orders);
                            return ListView.builder(
                              padding: const EdgeInsets.all(20),
                              itemCount: groupedOrders.length,
                              itemBuilder: (context, index) {
                                final group = groupedOrders[index];
                                return Padding(
                                  padding: EdgeInsets.only(
                                    bottom: index < groupedOrders.length - 1
                                        ? 16
                                        : 0,
                                  ),
                                  child: _buildOrderCard(
                                    group['order'] as OrderEntity,
                                    group['count'] as int,
                                  ),
                                );
                              },
                            );
                          },
                        ),
                      ),
              ),
            ],
          ),
        ),
      ),
    );
  }

  bool _canReturnOrder(String status) {
    return status.toLowerCase() == 'pending' || 
           status.toLowerCase() == 'in_progress';
  }

  Future<void> _showReturnOrderDialog(OrderEntity order) async {
    final confirmed = await showDialog<bool>(
      context: context,
      builder: (context) => AlertDialog(
        title: const Text('Отменить заказ?'),
        content: const Text(
          'Вы уверены, что хотите отменить заказ?',
        ),
        actions: [
          TextButton(
            onPressed: () => Navigator.of(context).pop(false),
            child: const Text('Отмена'),
          ),
          ElevatedButton(
            onPressed: () => Navigator.of(context).pop(true),
            style: ElevatedButton.styleFrom(
              backgroundColor: AppTheme.errorColor,
            ),
            child: const Text('Подтвердить'),
          ),
        ],
      ),
    );

    if (confirmed == true && mounted) {
      await _returnOrder(order.id);
    }
  }

  Future<void> _returnOrder(String orderId) async {
    try {
      await _di.returnOrderUseCase(orderId);
      
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          const SnackBar(
            content: Text('Заказ отменён'),
            backgroundColor: AppTheme.successColor,
          ),
        );
        await _loadOrders();
      }
    } catch (e) {
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(
            content: Text('Ошибка: ${e.toString()}'),
            backgroundColor: AppTheme.errorColor,
          ),
        );
      }
    }
  }

  Widget _buildOrderCard(OrderEntity order, int count) {
    final statusColor = _getStatusColor(order.status);
    final statusIcon = _getStatusIcon(order.status);
    final canReturn = _canReturnOrder(order.status);

    return GlassmorphicCard(
      padding: const EdgeInsets.all(16),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          // Status Badge
          Row(
            children: [
              Container(
                padding: const EdgeInsets.symmetric(
                  horizontal: 12,
                  vertical: 6,
                ),
                decoration: BoxDecoration(
                  color: statusColor.withValues(alpha: 0.2),
                  borderRadius: BorderRadius.circular(20),
                  border: Border.all(color: statusColor, width: 1),
                ),
                child: Row(
                  mainAxisSize: MainAxisSize.min,
                  children: [
                    Icon(statusIcon, size: 14, color: statusColor),
                    const SizedBox(width: 6),
                    Text(
                      order.statusRu,
                      style: TextStyle(
                        color: statusColor,
                        fontSize: 12,
                        fontWeight: FontWeight.w600,
                      ),
                    ),
                  ],
                ),
              ),
              const Spacer(),
              Text(
                _formatDate(order.createdAt),
                style: Theme.of(
                  context,
                ).textTheme.bodySmall?.copyWith(color: AppTheme.textSecondary),
              ),
            ],
          ),

          const SizedBox(height: 12),

          // Order ID
          Row(
            children: [
              Icon(
                FontAwesomeIcons.hashtag,
                size: 14,
                color: AppTheme.textSecondary,
              ),
              const SizedBox(width: 6),
              Text(
                'Заказ ${order.id.length > 8 ? order.id.substring(0, 8) : order.id}',
                style: Theme.of(
                  context,
                ).textTheme.titleMedium?.copyWith(fontWeight: FontWeight.w600),
              ),
            ],
          ),

          const SizedBox(height: 8),

          // Good Name with Quantity
          Row(
            children: [
              Icon(
                FontAwesomeIcons.boxOpen,
                size: 14,
                color: AppTheme.textSecondary,
              ),
              const SizedBox(width: 6),
              Expanded(
                child: Text(
                  '${order.good?.name ?? 'Товар #${order.goodId.substring(0, 8)}'} × $count',
                  style: Theme.of(context).textTheme.bodyMedium?.copyWith(
                    color: AppTheme.textSecondary,
                  ),
                  overflow: TextOverflow.ellipsis,
                ),
              ),
            ],
          ),

          // Return button
          if (canReturn) ...[
            const SizedBox(height: 12),
            SizedBox(
              width: double.infinity,
              child: ElevatedButton.icon(
                onPressed: () => _showReturnOrderDialog(order),
                style: ElevatedButton.styleFrom(
                  backgroundColor: AppTheme.errorColor,
                  foregroundColor: Colors.white,
                  padding: const EdgeInsets.symmetric(vertical: 12),
                  shape: RoundedRectangleBorder(
                    borderRadius: BorderRadius.circular(12),
                  ),
                ),
                icon: const Icon(FontAwesomeIcons.ban, size: 16),
                label: const Text(
                  'Отменить заказ',
                  style: TextStyle(fontWeight: FontWeight.w600),
                ),
              ),
            ),
          ],
        ],
      ),
    );
  }

  String _formatDate(DateTime date) {
    final now = DateTime.now();
    final difference = now.difference(date);

    if (difference.inMinutes < 60) {
      return '${difference.inMinutes} мин назад';
    } else if (difference.inHours < 24) {
      return '${difference.inHours} ч назад';
    } else {
      return '${difference.inDays} д назад';
    }
  }
}
