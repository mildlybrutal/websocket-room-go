import { useState, useEffect } from "react";
import { Link, useLocation } from "react-router-dom";
import { Hash, LogOut, Plus, Search, MessageSquare, Settings } from "lucide-react";
import { useAuth } from "../../context/AuthContext";
import { api } from "../../services/api";
import type { Room } from "../../types/room";
import { CreateRoomModal } from "../Modals/CreateRoomModal";
import { JoinRoomModal } from "../Modals/JoinRoomModal";

export function Sidebar() {
    const { user, logout, token } = useAuth();
    const [rooms, setRooms] = useState<Room[]>([]);
    const [isCreateOpen, setIsCreateOpen] = useState(false);
    const [isJoinOpen, setIsJoinOpen] = useState(false);
    const location = useLocation();

    useEffect(() => {
        if (token) {
            fetchRooms();
        }
    }, [token]);

    const fetchRooms = async () => {
        try {
            if (token) {
                const data = await api.getRooms(token);
                setRooms(data);
            }
        } catch (error) {
            console.error("Failed to fetch rooms", error);
        }
    };

    const handleRoomCreated = (name: string) => {
        fetchRooms();
    };

    const handleRoomJoined = (room: Room) => {
        fetchRooms();
    };

    return (
        <>
            <aside className="w-64 bg-cat-mantle border-r border-cat-surface0 flex flex-col h-full">
                {/* User Header */}
                <div className="p-4 border-b border-cat-surface0 flex items-center justify-between">
                    <div className="flex items-center space-x-3">
                        <div className="w-10 h-10 rounded-full bg-gradient-to-tr from-cat-blue to-cat-mauve flex items-center justify-center font-bold text-lg text-cat-base shadow-lg">
                            {user?.username?.charAt(0).toUpperCase()}
                        </div>
                        <div className="overflow-hidden">
                            <h3 className="font-bold text-cat-text truncate">{user?.username}</h3>
                            <span className="text-xs text-cat-green flex items-center font-medium">
                                <span className="w-2 h-2 bg-cat-green rounded-full mr-1 animate-pulse"></span>
                                Online
                            </span>
                        </div>
                    </div>
                </div>

                {/* Navigation/Channels */}
                <div className="flex-1 overflow-y-auto py-4 custom-scrollbar">
                    <div className="px-4 mb-2 flex items-center justify-between">
                        <div className="text-xs font-bold text-cat-overlay0 uppercase tracking-wider">
                            Rooms
                        </div>
                        <div className="flex space-x-1">
                            <button
                                onClick={() => setIsJoinOpen(true)}
                                className="p-1 text-cat-overlay0 hover:text-cat-blue hover:bg-cat-surface0 rounded-lg transition-all"
                                title="Join Room"
                            >
                                <Search size={14} />
                            </button>
                            <button
                                onClick={() => setIsCreateOpen(true)}
                                className="p-1 text-cat-overlay0 hover:text-cat-green hover:bg-cat-surface0 rounded-lg transition-all"
                                title="Create Room"
                            >
                                <Plus size={14} />
                            </button>
                        </div>
                    </div>

                    <nav className="space-y-0.5 px-2">
                        {rooms.map((room) => {
                            const isActive = location.pathname === `/room/${room.id}`;
                            return (
                                <Link
                                    key={room.id}
                                    to={`/room/${room.id}`}
                                    className={`flex items-center px-3 py-2 rounded-lg transition-all group ${isActive
                                        ? "bg-cat-surface0 text-cat-text font-medium"
                                        : "text-cat-subtext0 hover:bg-cat-surface0/50 hover:text-cat-text"
                                        }`}
                                >
                                    <Hash
                                        size={18}
                                        className={`mr-3 transition-colors ${isActive
                                            ? "text-cat-mauve"
                                            : "text-cat-overlay0 group-hover:text-cat-mauve"
                                            }`}
                                    />
                                    <span className="truncate">{room.name}</span>
                                    {isActive && (
                                        <div className="ml-auto w-1.5 h-1.5 bg-cat-mauve rounded-full"></div>
                                    )}
                                </Link>
                            );
                        })}
                    </nav>

                    {/* Direct Messages Placeholder */}
                    <div className="px-4 mt-8 mb-2 flex items-center justify-between">
                        <div className="text-xs font-bold text-cat-overlay0 uppercase tracking-wider">
                            Direct Messages
                        </div>
                        <button className="p-1 text-cat-overlay0 hover:text-cat-text hover:bg-cat-surface0 rounded-lg transition-all">
                            <Plus size={14} />
                        </button>
                    </div>
                    <div className="px-4 py-8 text-center">
                        <MessageSquare className="w-8 h-8 text-cat-surface1 mx-auto mb-2" />
                        <p className="text-xs text-cat-overlay0">No conversations yet</p>
                    </div>
                </div>

                {/* Footer Actions */}
                <div className="p-4 border-t border-cat-surface0 space-y-2">
                    <button className="flex items-center w-full px-3 py-2 text-cat-subtext0 hover:bg-cat-surface0 hover:text-cat-text rounded-lg transition-colors text-sm font-medium">
                        <Settings size={18} className="mr-3" />
                        Settings
                    </button>
                    <button
                        onClick={logout}
                        className="flex items-center w-full px-3 py-2 text-cat-red hover:bg-cat-red/10 rounded-lg transition-colors text-sm font-medium"
                    >
                        <LogOut size={18} className="mr-3" />
                        Sign Out
                    </button>
                </div>
            </aside>

            <CreateRoomModal
                isOpen={isCreateOpen}
                onClose={() => setIsCreateOpen(false)}
                onSuccess={handleRoomCreated}
            />

            <JoinRoomModal
                isOpen={isJoinOpen}
                onClose={() => setIsJoinOpen(false)}
                onSuccess={handleRoomJoined}
            />
        </>
    );
}
