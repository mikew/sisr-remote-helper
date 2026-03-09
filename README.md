# sisr-remote-helper

Helper to launch a UWP app from Steam and start / stop SISR alongside.

## Usage

sisr-remote-helper is designed to be added directly to Steam.

1. Download the latest release for your platform
1. Place the executable next to the SISR executable

Now you can start adding sisr-remote-helper to Steam and launch your UWP /
finicky third-party launcher games directly.

### UWP

1. In Steam, add a Non-Steam game to your library and select `sisr-remote-helper.exe`
1. In the properties of the newly added game, set the Launch Options to:

   ```
   uwp AUMID
   ```

   IE for Minecraft:

   ```
   uwp MICROSOFT.MINECRAFTUWP_8wekyb3d8bbwe!Game
   ```

To list your UWP apps and their AUMIDs, run:

```
sisr-remote-helper listapps
```

This will open a new window with the list.

### win32

1. In Steam, add a Non-Steam game to your library and select `sisr-remote-helper.exe`
1. In the properties of the newly added game, set the Launch Options to:

   ```
   win32 EXE_PATH
   ```

   IE for The Rogue Prince of Persia from Ubisoft Connect:

   ```
   win32 --grep upc.exe "C:\Program Files (x86)\Ubisoft\Ubisoft Game Launcher\games\TheRoguePrinceOfPersia\The Rogue Prince of Persia.exe"
   ```

Using `--grep ...` will keep sisr-remote-helper running as long as any process
running matches.

So for the Price of Persia example, when launching `The Rogue Prince of
Persia.exe` it exits immediately to launch Ubisoft Connect, which then launches
the game again. With `--grep upc.exe`, sisr-remote-helper will keep running
until Ubisoft Connect exits.

### More Info

You can also specify a path to a SISR config, and stop SISR from starting altogether.

Refer to the help message for more info:

```
NAME:
   sisr-remote-helper uwp - Launch SISR and a UWP app

USAGE:
   sisr-remote-helper uwp [options] <aumid>

OPTIONS:
   --[no-]start-sisr                Whether to start SISR automatically (default: true)
   --sisr-path string               (default: "./SISR")
   --sisr-config string
   --grep string [ --grep string ]  Also consider the app running if any process exe path contains this string (can be specified multiple times)
   --help, -h                       show help
```

```
NAME:
   sisr-remote-helper win32 - Launch SISR and a Win32 executable

USAGE:
   sisr-remote-helper win32 [options] <exe-path>

OPTIONS:
   --[no-]start-sisr                Whether to start SISR automatically (default: true)
   --sisr-path string               (default: "./SISR")
   --sisr-config string
   --grep string [ --grep string ]  Also consider the app running if any process exe path contains this string (can be specified multiple times)
   --help, -h                       show help
```
