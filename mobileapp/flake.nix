{
  description = "Flutter + Android dev env for SkyPostDelivery";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
    flake-utils.url = "github:numtide/flake-utils";
  };

  outputs = { self, nixpkgs, flake-utils }:
    flake-utils.lib.eachSystem [ "x86_64-linux" ] (system:
      let
        pkgs = import nixpkgs {
          inherit system;
          config = {
            allowUnfree = true;
            android_sdk.accept_license = true;
          };
        };

        androidEnv = pkgs.androidenv.override {
          licenseAccepted = true;
        };

        androidComposition = androidEnv.composeAndroidPackages {
          platformVersions = [
            "34"
            "36"
          ];

          buildToolsVersions = [
            "34.0.0"
            "35.0.0"
          ];

          includeNDK = true;
          ndkVersions = [
            "28.2.13676358"
            "29.0.14206865"
          ];

          cmakeVersions = [ "3.22.1" ];

          includeEmulator = false;
          includeSources = false;
          includeSystemImages = false;
        };

        androidSdk = androidComposition.androidsdk;
        androidHome = "${androidSdk}/libexec/android-sdk";
      in {
        devShells.default = pkgs.mkShell {
          ANDROID_HOME = androidHome;
          ANDROID_SDK_ROOT = androidHome;
          JAVA_HOME = pkgs.jdk17;

          buildInputs = with pkgs; [
            flutter
            androidSdk
            jdk17
            cmake
            ninja
            pkg-config
          ];

          shellHook = ''
            echo "ANDROID_HOME=$ANDROID_HOME"
            echo "JAVA_HOME=$JAVA_HOME"

            if [ -d "$ANDROID_HOME/ndk" ]; then
              export ANDROID_NDK_ROOT="$ANDROID_HOME/ndk/$(ls "$ANDROID_HOME/ndk" | head -n1)"
              echo "ANDROID_NDK_ROOT=$ANDROID_NDK_ROOT"
            fi

            echo "Platforms:"
            ls "$ANDROID_HOME/platforms" || true
            echo "Build-tools:"
            ls "$ANDROID_HOME/build-tools" || true
          '';
        };
      }
    );

}
