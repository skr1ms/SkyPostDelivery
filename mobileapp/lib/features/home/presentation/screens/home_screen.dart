import 'dart:async';
import 'package:flutter/material.dart';
import 'package:provider/provider.dart';
import 'package:font_awesome_flutter/font_awesome_flutter.dart';
import '../../../../core/theme/app_theme.dart';
import '../../../../core/utils/error_handler.dart';
import '../../../../core/di/injection_container.dart';
import '../../../auth/presentation/providers/auth_provider.dart';
import '../../../goods/domain/entities/good_entity.dart';
import '../../../auth/presentation/widgets/animated_background.dart';
import '../../../auth/presentation/widgets/glassmorphic_card.dart';

class HomeScreen extends StatefulWidget {
  final Function(Function(bool))? onVisibilityListenerReady;

  const HomeScreen({super.key, this.onVisibilityListenerReady});

  @override
  State<HomeScreen> createState() => _HomeScreenState();
}

class _HomeScreenState extends State<HomeScreen>
    with WidgetsBindingObserver, AutomaticKeepAliveClientMixin {
  List<GoodEntity> _goods = [];
  final Map<String, int> _cart = {};
  bool _isLoading = true;
  String? _error;
  Timer? _periodicTimer;
  bool _isScreenVisible = true;

  final _di = InjectionContainer();

  @override
  bool get wantKeepAlive => true;

  @override
  void initState() {
    super.initState();
    WidgetsBinding.instance.addObserver(this);
    widget.onVisibilityListenerReady?.call(onScreenVisibilityChanged);
    _loadGoods();
    _startPeriodicRefresh();
  }

  @override
  void dispose() {
    WidgetsBinding.instance.removeObserver(this);
    _periodicTimer?.cancel();
    super.dispose();
  }

  @override
  void didChangeAppLifecycleState(AppLifecycleState state) {
    super.didChangeAppLifecycleState(state);
    if (state == AppLifecycleState.resumed && _isScreenVisible) {
      _loadGoods(silent: true);
    }
  }

  void _startPeriodicRefresh() {
    _periodicTimer?.cancel();
    _periodicTimer = Timer.periodic(const Duration(minutes: 1), (timer) {
      if (_isScreenVisible && mounted) {
        _loadGoods(silent: true);
      }
    });
  }

  void onScreenVisibilityChanged(bool isVisible) {
    _isScreenVisible = isVisible;
    if (isVisible && mounted) {
      _loadGoods(silent: true);
    }
  }

  Future<void> _loadGoods({bool silent = false}) async {
    if (!silent) {
      setState(() {
        _isLoading = true;
        _error = null;
      });
    }

    try {
      final goods = await _di.getGoodsUseCase();

      if (mounted) {
        setState(() {
          _goods = goods;
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

  void _toggleGoodInCart(String goodId) {
    setState(() {
      if (_cart.containsKey(goodId)) {
        _cart.remove(goodId);
      } else {
        _cart[goodId] = 1;
      }
    });
  }

  void _updateQuantity(String goodId, int quantity) {
    final good = _goods.firstWhere((g) => g.id == goodId);

    if (quantity <= 0) {
      setState(() {
        _cart.remove(goodId);
      });
    } else if (quantity <= good.quantityAvailable) {
      setState(() {
        _cart[goodId] = quantity;
      });
    } else {
      ErrorHandler.showError(
        context,
        'Максимальное количество: ${good.quantityAvailable}',
      );
    }
  }

  void _showQuantityDialog(String goodId) {
    final good = _goods.firstWhere((g) => g.id == goodId);
    final currentQuantity = _cart[goodId] ?? 1;

    showDialog(
      context: context,
      builder: (context) => AlertDialog(
        backgroundColor: AppTheme.cardColor,
        title: Text(good.name, style: const TextStyle(color: Colors.white)),
        content: Column(
          mainAxisSize: MainAxisSize.min,
          children: [
            Text(
              'Доступно: ${good.quantityAvailable} шт',
              style: TextStyle(color: AppTheme.textSecondary, fontSize: 14),
            ),
            const SizedBox(height: 20),
            Row(
              mainAxisAlignment: MainAxisAlignment.center,
              children: [
                IconButton(
                  onPressed: currentQuantity > 1
                      ? () {
                          _updateQuantity(goodId, currentQuantity - 1);
                          Navigator.pop(context);
                          _showQuantityDialog(goodId);
                        }
                      : null,
                  icon: const Icon(Icons.remove_circle_outline),
                  color: AppTheme.primaryColor,
                  iconSize: 32,
                ),
                Container(
                  padding: const EdgeInsets.symmetric(
                    horizontal: 24,
                    vertical: 12,
                  ),
                  decoration: BoxDecoration(
                    color: AppTheme.primaryColor.withValues(alpha: 0.1),
                    borderRadius: BorderRadius.circular(12),
                  ),
                  child: Text(
                    '$currentQuantity',
                    style: const TextStyle(
                      color: AppTheme.primaryColor,
                      fontSize: 24,
                      fontWeight: FontWeight.bold,
                    ),
                  ),
                ),
                IconButton(
                  onPressed: currentQuantity < good.quantityAvailable
                      ? () {
                          _updateQuantity(goodId, currentQuantity + 1);
                          Navigator.pop(context);
                          _showQuantityDialog(goodId);
                        }
                      : null,
                  icon: const Icon(Icons.add_circle_outline),
                  color: AppTheme.primaryColor,
                  iconSize: 32,
                ),
              ],
            ),
          ],
        ),
        actions: [
          TextButton(
            onPressed: () {
              _cart.remove(goodId);
              Navigator.pop(context);
              setState(() {});
            },
            child: const Text(
              'Удалить',
              style: TextStyle(color: AppTheme.errorColor),
            ),
          ),
          ElevatedButton(
            onPressed: () => Navigator.pop(context),
            style: ElevatedButton.styleFrom(
              backgroundColor: AppTheme.primaryColor,
              foregroundColor: Colors.white,
            ),
            child: const Text('Готово'),
          ),
        ],
      ),
    );
  }

  Future<void> _createOrders() async {
    if (_cart.isEmpty) {
      ErrorHandler.showInfo(context, 'Корзина пуста');
      return;
    }

    final authProvider = Provider.of<AuthProvider>(context, listen: false);
    if (authProvider.user == null) return;

    try {
      showDialog(
        context: context,
        barrierDismissible: false,
        builder: (context) => const Center(child: CircularProgressIndicator()),
      );

      final List<String> goodIds = [];
      for (final entry in _cart.entries) {
        final goodId = entry.key;
        final quantity = entry.value;
        for (int i = 0; i < quantity; i++) {
          goodIds.add(goodId);
        }
      }

      await _di.createMultipleOrdersUseCase(
        userId: authProvider.user!.id,
        goodIds: goodIds,
      );

      if (!mounted) return;
      Navigator.of(context).pop();

      ErrorHandler.showSuccess(context, 'Создано заказов: ${goodIds.length}');

      setState(() {
        _cart.clear();
      });

      await _loadGoods();
    } catch (e) {
      if (mounted) {
        Navigator.of(context).pop();
        ErrorHandler.showError(context, e, onRetry: _createOrders);
      }
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
                      FontAwesomeIcons.store,
                      color: AppTheme.primaryColor,
                      size: 28,
                    ),
                    const SizedBox(width: 12),
                    Text(
                      'Каталог товаров',
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
                            const SizedBox(height: 8),
                            Text(
                              _error!,
                              style: Theme.of(context).textTheme.bodyMedium,
                              textAlign: TextAlign.center,
                            ),
                            const SizedBox(height: 24),
                            ElevatedButton.icon(
                              onPressed: _loadGoods,
                              icon: const Icon(FontAwesomeIcons.rotate),
                              label: const Text('Повторить'),
                            ),
                          ],
                        ),
                      )
                    : RefreshIndicator(
                        onRefresh: _loadGoods,
                        child: ListView.builder(
                          padding: const EdgeInsets.symmetric(horizontal: 20),
                          itemCount: _goods.length,
                          itemBuilder: (context, index) {
                            final good = _goods[index];
                            final isInCart = _cart.containsKey(good.id);
                            final quantity = _cart[good.id] ?? 0;

                            return _buildGoodCard(good, isInCart, quantity);
                          },
                        ),
                      ),
              ),

              if (_cart.isNotEmpty)
                GlassmorphicCard(
                  margin: const EdgeInsets.all(20),
                  padding: const EdgeInsets.all(16),
                  child: Row(
                    children: [
                      Expanded(
                        child: Column(
                          crossAxisAlignment: CrossAxisAlignment.start,
                          mainAxisSize: MainAxisSize.min,
                          children: [
                            Text(
                              'В корзине',
                              style: Theme.of(context).textTheme.bodySmall
                                  ?.copyWith(color: AppTheme.textSecondary),
                            ),
                            const SizedBox(height: 4),
                            Text(
                              '${_cart.length} ${_getItemsWord(_cart.length)}, '
                              '${_cart.values.reduce((a, b) => a + b)} шт',
                              style: Theme.of(context).textTheme.titleMedium
                                  ?.copyWith(
                                    color: AppTheme.primaryColor,
                                    fontWeight: FontWeight.w600,
                                  ),
                            ),
                          ],
                        ),
                      ),
                      const SizedBox(width: 16),
                      ElevatedButton.icon(
                        onPressed: _createOrders,
                        icon: const Icon(FontAwesomeIcons.cartShopping),
                        label: const Text('Заказать'),
                        style: ElevatedButton.styleFrom(
                          padding: const EdgeInsets.symmetric(
                            horizontal: 24,
                            vertical: 16,
                          ),
                          backgroundColor: AppTheme.primaryColor,
                          foregroundColor: Colors.white,
                        ),
                      ),
                    ],
                  ),
                ),
            ],
          ),
        ),
      ),
    );
  }

  String _getItemsWord(int count) {
    if (count % 10 == 1 && count % 100 != 11) return 'товар';
    if ([2, 3, 4].contains(count % 10) && ![12, 13, 14].contains(count % 100)) {
      return 'товара';
    }
    return 'товаров';
  }

  Widget _buildGoodCard(GoodEntity good, bool isInCart, int quantity) {
    return GestureDetector(
      onTap: () {
        if (!good.isAvailable) return;

        if (isInCart) {
          _showQuantityDialog(good.id);
        } else {
          _toggleGoodInCart(good.id);
        }
      },
      child: AnimatedContainer(
        duration: const Duration(milliseconds: 300),
        margin: const EdgeInsets.only(bottom: 16),
        child: GlassmorphicCard(
          margin: EdgeInsets.zero,
          padding: const EdgeInsets.all(16),
          child: Container(
            decoration: BoxDecoration(
              borderRadius: BorderRadius.circular(16),
              border: Border.all(
                color: isInCart ? AppTheme.primaryColor : Colors.transparent,
                width: 3,
              ),
            ),
            padding: const EdgeInsets.all(16),
            child: Row(
              children: [
                // Product Info
                Expanded(
                  child: Column(
                    crossAxisAlignment: CrossAxisAlignment.start,
                    children: [
                      Row(
                        children: [
                          Expanded(
                            child: Text(
                              good.name,
                              style: Theme.of(context).textTheme.titleMedium
                                  ?.copyWith(
                                    fontWeight: FontWeight.w600,
                                    color: isInCart
                                        ? AppTheme.primaryColor
                                        : Colors.white,
                                  ),
                            ),
                          ),
                          Container(
                            padding: const EdgeInsets.symmetric(
                              horizontal: 8,
                              vertical: 4,
                            ),
                            decoration: BoxDecoration(
                              color: good.isAvailable
                                  ? AppTheme.successColor.withValues(alpha: 0.2)
                                  : AppTheme.errorColor.withValues(alpha: 0.2),
                              borderRadius: BorderRadius.circular(8),
                              border: Border.all(
                                color: good.isAvailable
                                    ? AppTheme.successColor
                                    : AppTheme.errorColor,
                                width: 1,
                              ),
                            ),
                            child: Text(
                              good.isAvailable
                                  ? 'В наличии: ${good.quantityAvailable} шт'
                                  : 'Нет в наличии',
                              style: Theme.of(context).textTheme.bodySmall
                                  ?.copyWith(
                                    color: good.isAvailable
                                        ? AppTheme.successColor
                                        : AppTheme.errorColor,
                                    fontWeight: FontWeight.w600,
                                  ),
                            ),
                          ),
                        ],
                      ),
                      const SizedBox(height: 8),
                      Row(
                        children: [
                          Icon(
                            FontAwesomeIcons.boxOpen,
                            size: 14,
                            color: AppTheme.textSecondary,
                          ),
                          const SizedBox(width: 6),
                          Text(
                            good.dimensions,
                            style: Theme.of(context).textTheme.bodySmall
                                ?.copyWith(color: AppTheme.textSecondary),
                          ),
                          const SizedBox(width: 16),
                          Icon(
                            FontAwesomeIcons.weightHanging,
                            size: 14,
                            color: AppTheme.textSecondary,
                          ),
                          const SizedBox(width: 6),
                          Text(
                            '${good.weight} кг',
                            style: Theme.of(context).textTheme.bodySmall
                                ?.copyWith(color: AppTheme.textSecondary),
                          ),
                        ],
                      ),
                    ],
                  ),
                ),

                const SizedBox(width: 16),

                // Cart Icon with Quantity
                AnimatedContainer(
                  duration: const Duration(milliseconds: 200),
                  width: isInCart ? 60 : 40,
                  height: 40,
                  decoration: BoxDecoration(
                    color: good.isAvailable
                        ? (isInCart
                              ? AppTheme.primaryColor
                              : Colors.white.withValues(alpha: 0.1))
                        : Colors.grey.withValues(alpha: 0.1),
                    borderRadius: BorderRadius.circular(20),
                    border: Border.all(
                      color: good.isAvailable
                          ? (isInCart
                                ? AppTheme.primaryColor
                                : AppTheme.textSecondary.withValues(alpha: 0.5))
                          : Colors.grey,
                      width: 2,
                    ),
                    boxShadow: isInCart && good.isAvailable
                        ? [
                            BoxShadow(
                              color: AppTheme.primaryColor.withValues(
                                alpha: 0.3,
                              ),
                              blurRadius: 8,
                              spreadRadius: 2,
                            ),
                          ]
                        : null,
                  ),
                  child: isInCart
                      ? Row(
                          mainAxisAlignment: MainAxisAlignment.center,
                          children: [
                            const Icon(
                              FontAwesomeIcons.cartShopping,
                              size: 16,
                              color: Colors.white,
                            ),
                            const SizedBox(width: 4),
                            Text(
                              '$quantity',
                              style: const TextStyle(
                                color: Colors.white,
                                fontWeight: FontWeight.bold,
                                fontSize: 16,
                              ),
                            ),
                          ],
                        )
                      : Icon(
                          good.isAvailable
                              ? Icons.add_rounded
                              : Icons.close_rounded,
                          size: 24,
                          color: good.isAvailable
                              ? AppTheme.textSecondary
                              : Colors.grey,
                        ),
                ),
              ],
            ),
          ),
        ),
      ),
    );
  }
}
