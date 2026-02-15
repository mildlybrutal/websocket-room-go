import { useEffect, useRef, useState } from "react";
import { Send, Smile, Users } from "lucide-react";
import { useParams } from "react-router-dom";
import { MessageBubble } from "./MessageBubble";
import { useChatSocket } from "../../hooks/useChatSocket";
import { useAuth } from "../../context/AuthContext";
import { api } from "../../services/api";
import type { OnlineUser } from "../../types/chat";

export function ChatArea() {
    const { roomId = "1" } = useParams<{ roomId: string }>();
    const { token, user } = useAuth();
    const {
        messages,
        messageText,
        handleInput,
        sendMessage,
        connected,
        typingUsers,
        onlineCount,
        editMessage,
        deleteMessage,
    } = useChatSocket(roomId, token, user?.id);
    const bottomRef = useRef<HTMLDivElement>(null);
    const [globalOnline, setGlobalOnline] = useState<OnlineUser[]>([]);
    const [showOnlinePanel, setShowOnlinePanel] = useState(false);
    const [roomName, setRoomName] = useState<string>("");

    useEffect(() => {
        bottomRef.current?.scrollIntoView({ behavior: "smooth" });
    }, [messages]);

    // Fetch room name
    useEffect(() => {
        if (!token) return;
        const fetchRoomName = async () => {
            try {
                const rooms = await api.getRooms(token);
                const room = (Array.isArray(rooms) ? rooms : []).find(
                    (r: { id: number; name: string }) => String(r.id) === roomId
                );
                setRoomName(room?.name ?? `Room ${roomId}`);
            } catch {
                setRoomName(`Room ${roomId}`);
            }
        };
        fetchRoomName();
    }, [token, roomId]);

    // Poll global online users
    useEffect(() => {
        if (!token) return;
        const fetchOnline = async () => {
            try {
                const data = await api.getOnlineUsers(token);
                setGlobalOnline(data.online_users ?? []);
            } catch (e) {
                console.error("Failed to fetch online users", e);
            }
        };
        fetchOnline();
        const interval = setInterval(fetchOnline, 15000);
        return () => clearInterval(interval);
    }, [token]);

    return (
        <div className="flex-1 flex flex-col h-full bg-cat-base relative">
            {/* Header */}
            <div className="h-16 border-b border-cat-surface0 flex items-center justify-between px-6 bg-cat-base/95 backdrop-blur-md sticky top-0 z-10">
                <div className="flex items-center">
                    <h3 className="font-bold text-lg text-cat-text mr-3">
                        # {roomName}
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
                <button
                    onClick={() => setShowOnlinePanel(!showOnlinePanel)}
                    className="flex items-center gap-2 text-sm text-cat-subtext0 hover:text-cat-text transition-colors px-3 py-1.5 rounded-lg hover:bg-cat-surface0"
                >
                    <Users size={16} className="text-cat-green" />
                    <span className="font-medium">{globalOnline.length || onlineCount}</span>
                    <span className="text-xs">Online</span>
                </button>
            </div>

            <div className="flex flex-1 overflow-hidden">
                {/* Messages */}
                <div className="flex-1 overflow-y-auto p-6 space-y-1 custom-scrollbar">
                    {messages.length === 0 && (
                        <div className="flex items-center justify-center h-full">
                            <div className="text-center">
                                <div className="w-16 h-16 rounded-full bg-cat-surface0 flex items-center justify-center mx-auto mb-4">
                                    <Send size={24} className="text-cat-overlay0" />
                                </div>
                                <p className="text-cat-overlay0 text-sm">No messages yet. Start the conversation!</p>
                            </div>
                        </div>
                    )}
                    {messages.map((msg) => (
                        <MessageBubble
                            key={msg.id}
                            message={msg}
                            onEdit={editMessage}
                            onDelete={deleteMessage}
                        />
                    ))}
                    <div ref={bottomRef} />
                </div>

                {/* Online Users Panel */}
                {showOnlinePanel && (
                    <div className="w-56 border-l border-cat-surface0 bg-cat-mantle overflow-y-auto custom-scrollbar">
                        <div className="p-3 border-b border-cat-surface0">
                            <h4 className="text-xs font-bold text-cat-overlay0 uppercase tracking-wider">
                                Online — {globalOnline.length}
                            </h4>
                        </div>
                        <div className="p-2 space-y-0.5">
                            {globalOnline.map((u) => (
                                <div key={u.id} className="flex items-center gap-2 px-2 py-1.5 rounded-lg hover:bg-cat-surface0/50 transition-colors">
                                    <div className="relative">
                                        <div className="w-7 h-7 rounded-full bg-gradient-to-tr from-cat-blue to-cat-mauve flex items-center justify-center text-[10px] font-bold text-cat-base">
                                            {u.username.charAt(0).toUpperCase()}
                                        </div>
                                        <div className="absolute -bottom-0.5 -right-0.5 w-2.5 h-2.5 bg-cat-green rounded-full border-2 border-cat-mantle" />
                                    </div>
                                    <span className="text-sm text-cat-text truncate font-medium">{u.username}</span>
                                </div>
                            ))}
                            {globalOnline.length === 0 && (
                                <p className="text-xs text-cat-overlay0 text-center py-4">No users online</p>
                            )}
                        </div>
                    </div>
                )}
            </div>

            {/* Typing Indicator */}
            {typingUsers.length > 0 && (
                <div className="absolute bottom-24 left-6 text-xs text-cat-overlay0 animate-pulse bg-cat-base/50 px-2 py-1 rounded-md">
                    {typingUsers.join(", ")} is typing...
                </div>
            )}

            {/* Input Area */}
            <div className="p-4 bg-cat-base border-t border-cat-surface0">
                <form
                    className="flex items-center gap-2 max-w-4xl mx-auto bg-cat-surface0 p-1.5 pl-4 rounded-2xl border border-transparent focus-within:border-cat-mauve/40 focus-within:ring-1 focus-within:ring-cat-mauve/20 transition-all shadow-sm"
                    onSubmit={sendMessage}
                >
                    <input
                        type="text"
                        className="flex-1 bg-transparent border-none focus:ring-0 focus:outline-none text-cat-text placeholder-cat-overlay0 text-sm"
                        placeholder="Type a message..."
                        value={messageText}
                        onChange={(e) => handleInput(e.target.value)}
                    />

                    <button type="button" className="p-2 text-cat-overlay0 hover:text-cat-text transition-colors hover:bg-cat-surface1 rounded-xl">
                        <Smile size={18} />
                    </button>

                    <button
                        type="submit"
                        disabled={!messageText.trim()}
                        className="p-2 bg-cat-mauve hover:bg-cat-mauve/80 text-cat-crust rounded-xl transition-all disabled:opacity-30 disabled:cursor-not-allowed transform active:scale-95"
                    >
                        <Send size={16} />
                    </button>
                </form>
            </div>
        </div>
    );
}
