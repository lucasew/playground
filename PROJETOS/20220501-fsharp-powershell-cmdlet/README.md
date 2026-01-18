# Demo PowerShell module with F#

Based on: https://medium.com/@natelehman/writing-powershell-modules-in-f-ed52704d97ed

## Steps

- `dotnet new sln`
- `dotnet new classlib -lang 'F#' -o src/PSModule`
- `dotnet sln add src/PSModule/PSModule.fsproj`
- `dotnet add src/PSModule package PowerShellStandard.Library`
-
  - type the code *
- `dotnet build`
- `dotnet publish`
- `pwsh`
- `Import-Module ./src/PSModule/bin/Debug/net6.0/publish/PSModule.dll` -- It's expected to give an error saying it misses FSharp.Core
- `Add-Type -path ./src/PSModule/bin/Debug/net6.0/publish/FSharp.Core.dll` -- If the error happens
- `Get-Help Get-Hello` to show command help
