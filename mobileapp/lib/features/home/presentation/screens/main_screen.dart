import 'package:flutter/material.dart';
import 'package:font_awesome_flutter/font_awesome_flutter.dart';
import '../../../../core/theme/app_theme.dart';
import 'home_screen.dart';
import '../../../orders/presentation/screens/orders_screen.dart';
import '../../../qr/presentation/screens/qr_screen.dart';
import '../../../profile/presentation/screens/profile_screen.dart';

class MainScreen extends StatefulWidget {
  const MainScreen({super.key});

  @override
  State<MainScreen> createState() => _MainScreenState();
}

class _MainScreenState extends State<MainScreen> {
  int _currentIndex = 0;
  Function(bool)? _homeScreenVisibilityCallback;

  late final List<Widget> _screens = [
    HomeScreen(
      onVisibilityListenerReady: (callback) {
        _homeScreenVisibilityCallback = callback;
      },
    ),
    const OrdersScreen(),
    const QRScreen(),
    const ProfileScreen(),
  ];

  void _onTabChanged(int index) {
    setState(() {
      _currentIndex = index;
    });

    _homeScreenVisibilityCallback?.call(index == 0);
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      body: IndexedStack(index: _currentIndex, children: _screens),
      bottomNavigationBar: Container(
        decoration: BoxDecoration(
          gradient: LinearGradient(
            begin: Alignment.topCenter,
            end: Alignment.bottomCenter,
            colors: [
              AppTheme.surfaceColor.withValues(alpha: 0.95),
              AppTheme.cardColor.withValues(alpha: 0.98),
            ],
          ),
          boxShadow: [
            BoxShadow(
              color: Colors.black.withValues(alpha: 0.2),
              blurRadius: 20,
              offset: const Offset(0, -5),
            ),
          ],
        ),
        child: SafeArea(
          child: Padding(
            padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 8),
            child: Row(
              mainAxisAlignment: MainAxisAlignment.spaceAround,
              children: [
                _buildNavItem(
                  index: 0,
                  icon: FontAwesomeIcons.store,
                  label: 'Каталог',
                  color: AppTheme.primaryColor,
                ),
                _buildNavItem(
                  index: 1,
                  icon: FontAwesomeIcons.clipboardList,
                  label: 'Заказы',
                  color: AppTheme.accentColor,
                ),
                _buildNavItem(
                  index: 2,
                  icon: FontAwesomeIcons.qrcode,
                  label: 'QR-код',
                  color: AppTheme.successColor,
                ),
                _buildNavItem(
                  index: 3,
                  icon: FontAwesomeIcons.user,
                  label: 'Профиль',
                  color: AppTheme.secondaryColor,
                ),
              ],
            ),
          ),
        ),
      ),
    );
  }

  Widget _buildNavItem({
    required int index,
    required IconData icon,
    required String label,
    required Color color,
  }) {
    final isSelected = _currentIndex == index;

    return GestureDetector(
      onTap: () => _onTabChanged(index),
      child: AnimatedContainer(
        duration: const Duration(milliseconds: 300),
        curve: Curves.easeInOut,
        padding: EdgeInsets.symmetric(
          horizontal: isSelected ? 16 : 12,
          vertical: 8,
        ),
        decoration: BoxDecoration(
          gradient: isSelected
              ? LinearGradient(
                  colors: [
                    color.withValues(alpha: 0.2),
                    color.withValues(alpha: 0.1),
                  ],
                  begin: Alignment.topLeft,
                  end: Alignment.bottomRight,
                )
              : null,
          borderRadius: BorderRadius.circular(16),
          border: isSelected
              ? Border.all(color: color.withValues(alpha: 0.5), width: 1)
              : null,
        ),
        child: Row(
          mainAxisSize: MainAxisSize.min,
          children: [
            Icon(
              icon,
              color: isSelected ? color : AppTheme.textSecondary,
              size: isSelected ? 22 : 20,
            ),
            if (isSelected) ...[
              const SizedBox(width: 8),
              Text(
                label,
                style: TextStyle(
                  color: color,
                  fontSize: 14,
                  fontWeight: FontWeight.w600,
                ),
              ),
            ],
          ],
        ),
      ),
    );
  }
}
