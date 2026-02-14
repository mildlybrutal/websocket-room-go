import { useState, useEffect } from "react";
import { X, Loader2, Users } from "lucide-react";
import { api } from "../../services/api";
import { useAuth } from "../../context/AuthContext";
import type { Room } from "../../types/room";

interface JoinRoomModalProps {
    isOpen: boolean;
    onClose: () => void;
    onSuccess: (room: Room) => void;
}

export function JoinRoomModal({ isOpen, onClose, onSuccess }: JoinRoomModalProps) {
    const [rooms, setRooms] = useState<Room[]>([]);
    const [loading, setLoading] = useState(false);
    const [joiningId, setJoiningId] = useState<number | null>(null);
    const { token } = useAuth();
    const [error, setError] = useState<string | null>(null);

    useEffect(() => {
        if (isOpen && token) {
            fetchRooms();
        }
    }, [isOpen, token]);

    const fetchRooms = async () => {
        setLoading(true);
        setError(null);
        try {
            if (token) {
                const data = await api.getRooms(token);
                setRooms(data);
            }
        } catch (err) {
            setError("Failed to load rooms");
        } finally {
            setLoading(false);
        }
    };

    const handleJoin = async (roomId: number) => {
        setJoiningId(roomId);
        try {
            if (token) {
                const response = await api.joinRoom(roomId, token);
                onSuccess(response.room);
                onClose();
            }
        } catch (err: any) {
            // If already joined, just proceed (or show error? API returns success if redundant)
            // API will return success either way or error.
            setError(err.message || "Failed to join");
        } finally {
            setJoiningId(null);
        }
    };

    if (!isOpen) return null;

    return (
        <div className="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/50 backdrop-blur-sm">
            <div className="bg-cat-mantle border border-cat-surface0 rounded-2xl w-full max-w-md shadow-2xl animate-in fade-in zoom-in duration-200 flex flex-col max-h-[80vh]">
                <div className="flex items-center justify-between p-4 border-b border-cat-surface0">
                    <h2 className="text-xl font-bold text-cat-text">Join Room</h2>
                    <button
                        onClick={onClose}
                        className="text-cat-overlay0 hover:text-cat-red transition-colors"
                    >
                        <X size={20} />
                    </button>
                </div>

                <div className="p-4 flex-1 overflow-y-auto custom-scrollbar space-y-2">
                    {loading ? (
                        <div className="flex justify-center py-8">
                            <Loader2 className="animate-spin text-cat-blue" size={32} />
                        </div>
                    ) : error ? (
                        <div className="text-center text-cat-red py-4">{error}</div>
                    ) : rooms.length === 0 ? (
                        <div className="text-center text-cat-overlay0 py-8">
                            No rooms found. Create one!
                        </div>
                    ) : (
                        rooms.map((room) => (
                            <div
                                key={room.id}
                                className="flex items-center justify-between p-3 rounded-xl bg-cat-surface0/50 hover:bg-cat-surface0 transition-colors border border-transparent hover:border-cat-surface1 group"
                            >
                                <div className="flex items-center space-x-3">
                                    <div className="w-10 h-10 rounded-full bg-cat-surface1 flex items-center justify-center text-cat-blue group-hover:bg-cat-blue group-hover:text-cat-base transition-colors">
                                        <Users size={20} />
                                    </div>
                                    <div>
                                        <h3 className="font-semibold text-cat-text">{room.name}</h3>
                                        <p className="text-xs text-cat-overlay0">ID: {room.id}</p>
                                    </div>
                                </div>
                                <button
                                    onClick={() => handleJoin(room.id)}
                                    disabled={joiningId === room.id}
                                    className="px-3 py-1.5 bg-cat-surface1 hover:bg-cat-green text-cat-text hover:text-cat-base rounded-lg text-sm font-medium transition-colors disabled:opacity-50"
                                >
                                    {joiningId === room.id ? (
                                        <Loader2 size={16} className="animate-spin" />
                                    ) : (
                                        "Join"
                                    )}
                                </button>
                            </div>
                        ))
                    )}
                </div>
            </div>
        </div>
    );
}
