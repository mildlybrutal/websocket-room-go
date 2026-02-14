import { useEffect, useRef } from "react";
import { Send, Plus, Smile } from "lucide-react";
import { useParams } from "react-router-dom";
import { MessageBubble } from "./MessageBubble";
import { useChatSocket } from "../../hooks/useChatSocket";
import { useAuth } from "../../context/AuthContext";

export function ChatArea() {
    const { roomId = "1" } = useParams<{ roomId: string }>();
    const { token } = useAuth();
    const { messages, messageText, handleInput, sendMessage, connected, presence, onlineCount } = useChatSocket(roomId, token);
    const bottomRef = useRef<HTMLDivElement>(null);

    useEffect(() => {
        bottomRef.current?.scrollIntoView({ behavior: "smooth" });
    }, [messages]);

    return (
        <div className="flex-1 flex flex-col h-full bg-cat-base relative">
            {/* Header */}
            <div className="h-16 border-b border-cat-surface0 flex items-center justify-between px-6 bg-cat-base/95 backdrop-blur-md sticky top-0 z-10">
                <div className="flex items-center">
                    <h3 className="font-bold text-lg text-cat-text mr-3">
                        # {roomId === "1" ? "General" : roomId === "2" ? "Random" : roomId === "3" ? "Tech" : `Room ${roomId}`}
                    </h3>
                    {connected ? (
                        <span className="flex items-center text-xs text-cat-green bg-cat-green/10 px-2 py-0.5 rounded-full border border-cat-green/20">
                            Connected
                        </span>
                    ) : (
                        <span className="flex items-center text-xs text-cat-red bg-cat-red/10 px-2 py-0.5 rounded-full border border-cat-red/20">
                            Disconnected
                        </span>
                    )}
                </div>
                <div className="text-sm text-cat-subtext0">
                    {onlineCount} Online
                </div>
            </div>

            {/* Messages */}
            <div className="flex-1 overflow-y-auto p-6 space-y-4 custom-scrollbar">
                {messages.map((msg) => (
                    <MessageBubble key={msg.id} message={msg} />
                ))}
                <div ref={bottomRef} />
            </div>

            {/* Typing Indicator */}
            {presence.typingUsers.length > 0 && (
                <div className="absolute bottom-24 left-6 text-xs text-cat-overlay0 animate-pulse bg-cat-base/50 px-2 py-1 rounded-md">
                    {presence.typingUsers.join(", ")} is typing...
                </div>
            )}

            {/* Input Area */}
            <div className="p-4 bg-cat-base border-t border-cat-surface0">
                <form
                    className="flex items-center space-x-3 max-w-4xl mx-auto bg-cat-surface0 p-2 rounded-2xl border border-transparent focus-within:border-cat-blue focus-within:ring-1 focus-within:ring-cat-blue/20 transition-all shadow-sm"
                    onSubmit={sendMessage}
                >
                    <button type="button" className="p-2 text-cat-overlay0 hover:text-cat-text transition-colors hover:bg-cat-surface1 rounded-xl">
                        <Plus size={20} />
                    </button>

                    <input
                        type="text"
                        className="flex-1 bg-transparent border-none focus:ring-0 text-cat-text placeholder-cat-overlay0 px-2"
                        placeholder="Type a message..."
                        value={messageText}
                        onChange={(e) => handleInput(e.target.value)}
                    />

                    <button type="button" className="p-2 text-cat-overlay0 hover:text-cat-text transition-colors hover:bg-cat-surface1 rounded-xl">
                        <Smile size={20} />
                    </button>

                    <button
                        type="submit"
                        disabled={!messageText.trim()}
                        className="p-2 bg-cat-blue hover:bg-cat-blue/90 text-cat-base rounded-xl transition-all disabled:opacity-50 disabled:cursor-not-allowed transform active:scale-95 shadow-md hover:shadow-lg"
                    >
                        <Send size={18} />
                    </button>
                </form>
            </div>
        </div>
    );
}
