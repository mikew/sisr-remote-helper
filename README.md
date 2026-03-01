# sisr-remote-helper

Helper to launch a UWP app from Steam and start / stop SISR alongside.

## Usage

sisr-remote-helper is designed to be added directly to Steam.

1. Download the latest release for your platform
1. Move place the executable next to the SISR executable
1. In Steam, add a Non-Steam game to your library and select `sisr-remote-helper.exe`
1. In the properties of the newly added game, set the Launch Options to:

    ```
    uwp AUMID
    ```

    IE for Minecraft:

    ```
    uwp MICROSOFT.MINECRAFTUWP_8wekyb3d8bbwe!Game
    ```

### More Info

You can also specify a path to a SISR config, and stop SISR from starting altogether.

Refer to the help message for more info:

```
NAME:
   sisr-remote-helper uwp - Launch SISR and a UWP app

USAGE:
   sisr-remote-helper uwp [options] <aumid>

OPTIONS:
   --[no-]start-sisr     Whether to start SISR automatically (default: true)
   --sisr-path string    (default: "./SISR")
   --sisr-config string
   --help, -h            show help
```
