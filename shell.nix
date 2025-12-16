{ pkgs ? import <nixpkgs> {
    config = {
      allowUnfree = true;
      android_sdk.accept_license = true;
    };
  }
}:

pkgs.mkShell {
  packages = with pkgs; [
    bash
    go
    gopls
    gotools
    go-tools
    nodejs
    nodePackages.npm
    nodePackages.typescript
    nodePackages.typescript-language-server
    python311
    python311Packages.pip
    python311Packages.virtualenv
    python311Packages.websockets
    python311Packages.pydantic
    python311Packages.python-dotenv
    python311Packages.numpy
    python311Packages.pytest
    python311Packages.pytest-asyncio
    python311Packages.pytest-cov
    gnumake
    git
    docker
    docker-compose
    postgresql_17
    protobuf
    grpcurl
    sqlc
    expat
    util-linux
    libbsd
    fontconfig
    freetype
    flutter
    android-studio-full
    android-tools
    clang
    cmake
    ninja
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
    pkgs.opencv4
    pkgs.expat
    pkgs.util-linux
    pkgs.libbsd
  ];

  shellHook = ''
    export ANDROID_HOME="$HOME/Android/Sdk"
    export ANDROID_SDK_ROOT="$ANDROID_HOME"

    export QT_QPA_PLATFORM=minimal
    export QT_QPA_PLATFORM_PLUGIN_PATH="$ANDROID_HOME/emulator/lib64/qt/plugins"
    export QT_OPENGL=software
    export QT_SCALE_FACTOR=none
    export LD_LIBRARY_PATH="$ANDROID_HOME/emulator/lib64/qt/lib:$ANDROID_HOME/emulator/lib64/vulkan:$ANDROID_HOME/emulator/lib64/gles_swiftshader:$ANDROID_HOME/emulator/lib64:$LD_LIBRARY_PATH"

    export PKG_CONFIG_PATH="${pkgs.opencv4}/lib/pkgconfig:$PKG_CONFIG_PATH"
    export CGO_CFLAGS="-I${pkgs.opencv4}/include/opencv4"
    export CGO_LDFLAGS="-L${pkgs.opencv4}/lib -L${pkgs.expat}/lib -L${pkgs.util-linux.lib}/lib -L${pkgs.libbsd}/lib"

    echo "=== SkyPostDelivery Dev Shell ==="
    echo "Go: $(go version | cut -d' ' -f3)"
    echo "Node: $(node --version)"
    echo "Python: $(python3 --version | cut -d' ' -f2)"
    echo "OpenCV: ${pkgs.opencv4.version}"
    echo "Flutter: $(flutter --version 2>&1 | head -1)"
    echo "ANDROID_HOME: $ANDROID_HOME"
    echo "=================================="
  '';
}

