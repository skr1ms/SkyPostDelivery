{ pkgs ? import <nixpkgs> {
    config = {
      allowUnfree = true;
      android_sdk.accept_license = true;
    };
  }
}:

pkgs.mkShell {
  packages = with pkgs; [
    flutter
    android-studio-full
    android-tools  # adb, fastboot
    clang
    cmake
    ninja
    pkg-config
    gtk3
    pcre
    libepoxy
    libuuid
    xorg.libXdmcp
    python3Packages.libselinux
    libsepol
    libthai
    libdatrie
    libxkbcommon
    dbus
    at-spi2-core
    xorg.libXtst
    pcre2
    jdk11
  ];

  LD_LIBRARY_PATH = pkgs.lib.makeLibraryPath [
    pkgs.fontconfig.lib
    pkgs.sqlite.out
  ];

  shellHook = ''
    export ANDROID_HOME="$HOME/Android/Sdk"
    export ANDROID_SDK_ROOT="$ANDROID_HOME"

    export QT_QPA_PLATFORM=minimal
    export QT_QPA_PLATFORM_PLUGIN_PATH="$ANDROID_HOME/emulator/lib64/qt/plugins"
    export QT_OPENGL=software
    export QT_SCALE_FACTOR=none
    export LD_LIBRARY_PATH="$ANDROID_HOME/emulator/lib64/qt/lib:$ANDROID_HOME/emulator/lib64/vulkan:$ANDROID_HOME/emulator/lib64/gles_swiftshader:$ANDROID_HOME/emulator/lib64:$LD_LIBRARY_PATH"

    echo "Flutter Android dev-shell активен"
    echo "ANDROID_HOME=$ANDROID_HOME"
  '';
}

