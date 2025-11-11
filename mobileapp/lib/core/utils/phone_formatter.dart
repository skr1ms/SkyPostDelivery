import 'package:flutter/services.dart';

class PhoneInputFormatter extends TextInputFormatter {
  @override
  TextEditingValue formatEditUpdate(
    TextEditingValue oldValue,
    TextEditingValue newValue,
  ) {
    final text = newValue.text;

    if (text.isEmpty) {
      return newValue.copyWith(text: '');
    }

    final digitsOnly = text.replaceAll(RegExp(r'\D'), '');

    if (digitsOnly.isEmpty) {
      return newValue.copyWith(text: '');
    }

    String formatted = '+7';

    if (digitsOnly.length > 1) {
      formatted += ' (';
      formatted += digitsOnly.substring(
        1,
        digitsOnly.length > 4 ? 4 : digitsOnly.length,
      );

      if (digitsOnly.length >= 4) {
        formatted += ')';
      }

      if (digitsOnly.length > 4) {
        formatted += ' ';
        formatted += digitsOnly.substring(
          4,
          digitsOnly.length > 7 ? 7 : digitsOnly.length,
        );
      }

      if (digitsOnly.length > 7) {
        formatted += '-';
        formatted += digitsOnly.substring(
          7,
          digitsOnly.length > 9 ? 9 : digitsOnly.length,
        );
      }

      if (digitsOnly.length > 9) {
        formatted += '-';
        formatted += digitsOnly.substring(
          9,
          digitsOnly.length > 11 ? 11 : digitsOnly.length,
        );
      }
    }

    return TextEditingValue(
      text: formatted,
      selection: TextSelection.collapsed(offset: formatted.length),
    );
  }
}
