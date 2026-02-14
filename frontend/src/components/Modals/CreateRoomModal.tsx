import { useState } from "react";
import { X, Loader2 } from "lucide-react";
import { api } from "../../services/api";
import { useAuth } from "../../context/AuthContext";

interface CreateRoomModalProps {
    isOpen: boolean;
    onClose: () => void;
    onSuccess: (roomName: string) => void;
}

export function CreateRoomModal({ isOpen, onClose, onSuccess }: CreateRoomModalProps) {
    const [name, setName] = useState("");
    const [loading, setLoading] = useState(false);
    const [error, setError] = useState<string | null>(null);
    const { token } = useAuth();

    if (!isOpen) return null;

    const handleSubmit = async (e: React.FormEvent) => {
        e.preventDefault();
        setError(null);
        setLoading(true);

        try {
            if (token) {
                await api.createRoom(name, token);
                onSuccess(name);
                setName("");
                onClose();
            }
        } catch (err: any) {
            setError(err.message || "Failed to create room");
        } finally {
            setLoading(false);
        }
    };

    return (
        <div className="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/50 backdrop-blur-sm">
            <div className="bg-cat-mantle border border-cat-surface0 rounded-2xl w-full max-w-md shadow-2xl animate-in fade-in zoom-in duration-200">
                <div className="flex items-center justify-between p-4 border-b border-cat-surface0">
                    <h2 className="text-xl font-bold text-cat-text">Create New Room</h2>
                    <button
                        onClick={onClose}
                        className="text-cat-overlay0 hover:text-cat-red transition-colors"
                    >
                        <X size={20} />
                    </button>
                </div>

                <form onSubmit={handleSubmit} className="p-4 space-y-4">
                    {error && (
                        <div className="p-3 bg-cat-red/10 border border-cat-red/20 text-cat-red text-sm rounded-lg">
                            {error}
                        </div>
                    )}

                    <div className="space-y-1">
                        <label className="text-sm font-medium text-cat-subtext0">Room Name</label>
                        <input
                            type="text"
                            value={name}
                            onChange={(e) => setName(e.target.value)}
                            placeholder="e.g. Project Alpha"
                            className="w-full px-3 py-2 bg-cat-surface0 border border-transparent rounded-xl text-cat-text focus:border-cat-blue focus:ring-1 focus:ring-cat-blue outline-none transition-all placeholder-cat-overlay0"
                            autoFocus
                            required
                        />
                    </div>

                    <div className="flex justify-end space-x-3 pt-2">
                        <button
                            type="button"
                            onClick={onClose}
                            className="px-4 py-2 text-sm font-medium text-cat-subtext0 hover:text-cat-text transition-colors"
                        >
                            Cancel
                        </button>
                        <button
                            type="submit"
                            disabled={loading || !name.trim()}
                            className="px-4 py-2 bg-cat-blue hover:bg-cat-blue/90 text-cat-base text-sm font-bold rounded-xl transition-all disabled:opacity-50 disabled:cursor-not-allowed flex items-center"
                        >
                            {loading ? <Loader2 size={16} className="animate-spin mr-2" /> : null}
                            Create Room
                        </button>
                    </div>
                </form>
            </div>
        </div>
    );
}
