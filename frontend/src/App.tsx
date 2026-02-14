import "./App.css";
import { ChatPanel } from "./components/ChatPanel";
import { ConnectionPanel } from "./components/ConnectionPanel";
import { LeftPanel } from "./components/LeftPanel";
import { useChatSocket } from "./hooks/useChatSocket";

function App() {
    const chat = useChatSocket();

    return (
        <div className="app-shell">
            <LeftPanel />
            <ChatPanel
                roomId={chat.roomId}
                onlineCount={chat.onlineCount}
                connected={chat.connected}
                messages={chat.messages}
                typingUsers={chat.presence.typingUsers}
                messageText={chat.messageText}
                onInputChange={chat.handleInput}
                onSendMessage={chat.sendMessage}
            />
            <ConnectionPanel
                wsUrl={chat.wsUrl}
                setWsUrl={chat.setWsUrl}
                clientId={chat.clientId}
                setClientId={chat.setClientId}
                token={chat.token}
                setToken={chat.setToken}
                roomId={chat.roomId}
                setRoomId={chat.setRoomId}
                connected={chat.connected}
                onConnect={chat.connect}
                onJoinRoom={chat.joinRoom}
            />
        </div>
    );
}

export default App;
