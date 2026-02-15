import { useState, useRef, useEffect } from "react";
import { format } from "date-fns";
import { Pencil, Trash2, Check, X } from "lucide-react";
import type { ChatMessage } from "../../types/chat";

interface MessageBubbleProps {
    message: ChatMessage;
    onEdit?: (messageId: number, newContent: string) => void;
    onDelete?: (messageId: number) => void;
}

export function MessageBubble({ message, onEdit, onDelete }: MessageBubbleProps) {
    const isMine = message.mine;
    const isSystem = message.kind === "system";
    const [isEditing, setIsEditing] = useState(false);
    const [editText, setEditText] = useState(message.content);
    const editInputRef = useRef<HTMLInputElement>(null);

    useEffect(() => {
        if (isEditing && editInputRef.current) {
            editInputRef.current.focus();
        }
    }, [isEditing]);

    if (isSystem) {
        return (
            <div className="flex justify-center my-4">
                <span className="px-3 py-1 text-xs text-cat-overlay2 bg-cat-surface0/50 rounded-full border border-cat-surface1">
                    {message.content}
                </span>
            </div>
        );
    }

    if (message.isDeleted) {
        return (
            <div className={`flex ${isMine ? "justify-end" : "justify-start"} mb-4`}>
                {!isMine && (
                    <div className="w-8 h-8 rounded-full bg-cat-surface1 flex items-center justify-center font-bold text-xs mr-2 self-end mb-1 text-cat-overlay0">
                        {message.senderName?.charAt(0).toUpperCase()}
                    </div>
                )}
                <div
                    className={`max-w-[75%] px-4 py-3 rounded-2xl ${isMine ? "rounded-br-sm" : "rounded-bl-sm"
                        } bg-cat-surface0/50 border border-cat-surface1`}
                >
                    {!isMine && <div className="text-xs text-cat-overlay0 mb-1 font-medium">{message.senderName}</div>}
                    <p className="text-sm text-cat-overlay0 italic">This message was deleted</p>
                    <div className="text-[10px] mt-1 text-right text-cat-overlay0 opacity-50">
                        {format(new Date(message.timestamp * 1000), "h:mm a")}
                    </div>
                </div>
            </div>
        );
    }

    const handleEditSubmit = () => {
        if (editText.trim() && editText !== message.content && message.messageId && onEdit) {
            onEdit(message.messageId, editText.trim());
        }
        setIsEditing(false);
    };

    const handleEditKeyDown = (e: React.KeyboardEvent) => {
        if (e.key === "Enter") {
            e.preventDefault();
            handleEditSubmit();
        }
        if (e.key === "Escape") {
            setEditText(message.content);
            setIsEditing(false);
        }
    };

    return (
        <div className={`flex ${isMine ? "justify-end" : "justify-start"} mb-4 group`}>
            {!isMine && (
                <div className="w-8 h-8 rounded-full bg-gradient-to-tr from-cat-mauve to-cat-blue flex items-center justify-center font-bold text-xs mr-2 self-end mb-1 text-cat-base shadow-sm">
                    {message.senderName?.charAt(0).toUpperCase()}
                </div>
            )}

            <div className="relative flex items-center gap-1">
                {/* Action buttons for own messages — show on left side for "mine" */}
                {isMine && message.messageId && !isEditing && (
                    <div className="flex items-center gap-0.5 opacity-0 group-hover:opacity-100 transition-opacity duration-200 mr-1">
                        <button
                            onClick={() => {
                                setEditText(message.content);
                                setIsEditing(true);
                            }}
                            className="p-1.5 rounded-lg text-cat-overlay0 hover:text-cat-blue hover:bg-cat-surface0 transition-all"
                            title="Edit message"
                        >
                            <Pencil size={14} />
                        </button>
                        <button
                            onClick={() => message.messageId && onDelete?.(message.messageId)}
                            className="p-1.5 rounded-lg text-cat-overlay0 hover:text-cat-red hover:bg-cat-red/10 transition-all"
                            title="Delete message"
                        >
                            <Trash2 size={14} />
                        </button>
                    </div>
                )}

                <div
                    className={`max-w-[75%] px-4 py-3 shadow-sm backdrop-blur-sm transition-all ${isMine
                        ? "bg-cat-blue text-cat-base rounded-2xl rounded-br-sm"
                        : "bg-cat-surface0 text-cat-text rounded-2xl rounded-bl-sm border border-cat-surface1"
                        }`}
                >
                    {!isMine && <div className="text-xs text-cat-mauve mb-1 font-bold">{message.senderName}</div>}

                    {isEditing ? (
                        <div className="flex items-center gap-2">
                            <input
                                ref={editInputRef}
                                type="text"
                                value={editText}
                                onChange={(e) => setEditText(e.target.value)}
                                onKeyDown={handleEditKeyDown}
                                className="flex-1 bg-cat-base/20 border border-cat-surface1 rounded-lg px-2 py-1 text-sm focus:outline-none focus:border-cat-mauve text-cat-base placeholder-cat-overlay0"
                            />
                            <button
                                onClick={handleEditSubmit}
                                className="p-1 rounded text-cat-green hover:bg-cat-green/20 transition-colors"
                                title="Save"
                            >
                                <Check size={14} />
                            </button>
                            <button
                                onClick={() => {
                                    setEditText(message.content);
                                    setIsEditing(false);
                                }}
                                className="p-1 rounded text-cat-red hover:bg-cat-red/20 transition-colors"
                                title="Cancel"
                            >
                                <X size={14} />
                            </button>
                        </div>
                    ) : (
                        <p className="text-sm leading-relaxed whitespace-pre-wrap">{message.content}</p>
                    )}

                    <div className={`flex items-center gap-1.5 mt-1 justify-end ${isMine ? "text-cat-crust" : "text-cat-overlay1"}`}>
                        {message.isEdited && (
                            <span className="text-[10px] font-medium opacity-60 italic">edited</span>
                        )}
                        <span className="text-[10px] font-medium opacity-70">
                            {format(new Date(message.timestamp * 1000), "h:mm a")}
                        </span>
                    </div>
                </div>

                {/* Action buttons for other's messages — show on right side */}
                {!isMine && message.messageId && !isEditing && (
                    <div className="flex items-center gap-0.5 opacity-0 group-hover:opacity-100 transition-opacity duration-200 ml-1">
                        {/* Others can't edit/delete, but placeholder for future reactions etc */}
                    </div>
                )}
            </div>
        </div>
    );
}
