class PhoneUtils {
  static String cleanPhone(String phone) {
    final digitsOnly = phone.replaceAll(RegExp(r'\D'), '');

    if (digitsOnly.isEmpty) {
      return '';
    }

    if (digitsOnly.startsWith('7') && digitsOnly.length == 11) {
      return digitsOnly;
    }

    if (digitsOnly.startsWith('8') && digitsOnly.length == 11) {
      return '7${digitsOnly.substring(1)}';
    }

    if (digitsOnly.length == 10) {
      return '7$digitsOnly';
    }

    return digitsOnly;
  }
}
