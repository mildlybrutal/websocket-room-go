type ConnectionPanelProps = {
    wsUrl: string;
    setWsUrl: (value: string) => void;
    clientId: string;
    setClientId: (value: string) => void;
    token: string;
    setToken: (value: string) => void;
    roomId: string;
    setRoomId: (value: string) => void;
    connected: boolean;
    onConnect: () => void;
    onJoinRoom: () => void;
};

export function ConnectionPanel({
    wsUrl,
    setWsUrl,
    clientId,
    setClientId,
    token,
    setToken,
    roomId,
    setRoomId,
    connected,
    onConnect,
    onJoinRoom,
}: ConnectionPanelProps) {
    return (
        <aside className="right-panel">
            <h3>Connection</h3>
            <label>
                WebSocket URL
                <input
                    value={wsUrl}
                    onChange={(event) => setWsUrl(event.target.value)}
                />
            </label>
            <label>
                Client ID
                <input
                    value={clientId}
                    onChange={(event) => setClientId(event.target.value)}
                />
            </label>
            <label>
                JWT token (optional)
                <input
                    value={token}
                    onChange={(event) => setToken(event.target.value)}
                    placeholder="Only needed if require_auth=true"
                />
            </label>
            <label>
                Room ID
                <input
                    value={roomId}
                    onChange={(event) => setRoomId(event.target.value)}
                />
            </label>
            <div className="button-row">
                <button className="secondary" onClick={onConnect}>
                    {connected ? "Reconnect" : "Connect"}
                </button>
                <button
                    className="secondary"
                    onClick={onJoinRoom}
                    disabled={!connected}
                >
                    Join room
                </button>
            </div>
        </aside>
    );
}
