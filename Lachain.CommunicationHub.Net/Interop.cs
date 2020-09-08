// ReSharper disable InconsistentNaming

namespace Lachain.CommunicationHub.Net
{
    [SymbolName(nameof(StartHub))]
    public delegate int StartHub();

    [SymbolName(nameof(StartHubOnPort))]
    public unsafe delegate int StartHubOnPort();

    [SymbolName(nameof(StopHub))]
    public delegate bool StopHub();

    [SymbolName(nameof(LogLevel))]
    public unsafe delegate bool LogLevel(byte* str, int len);
}
