import { format } from "date-fns";
import type { ChatMessage } from "../../types/chat";

interface MessageBubbleProps {
    message: ChatMessage;
}

export function MessageBubble({ message }: MessageBubbleProps) {
    const isMine = message.mine;
    const isSystem = message.kind === "system";

    if (isSystem) {
        return (
            <div className="flex justify-center my-4">
                <span className="px-3 py-1 text-xs text-cat-overlay2 bg-cat-surface0/50 rounded-full border border-cat-surface1">
                    {message.content}
                </span>
            </div>
        );
    }

    return (
        <div className={`flex ${isMine ? "justify-end" : "justify-start"} mb-4 group`}>
            {!isMine && (
                <div className="w-8 h-8 rounded-full bg-gradient-to-tr from-cat-mauve to-cat-blue flex items-center justify-center font-bold text-xs mr-2 self-end mb-1 text-cat-base shadow-sm">
                    {message.senderName?.charAt(0).toUpperCase()}
                </div>
            )}
            <div
                className={`max-w-[75%] px-4 py-3 shadow-sm backdrop-blur-sm transition-all ${isMine
                    ? "bg-cat-blue text-cat-base rounded-2xl rounded-br-sm"
                    : "bg-cat-surface0 text-cat-text rounded-2xl rounded-bl-sm border border-cat-surface1"
                    }`}
            >
                {!isMine && <div className="text-xs text-cat-mauve mb-1 font-bold">{message.senderName}</div>}
                <p className="text-sm leading-relaxed whitespace-pre-wrap">{message.content}</p>
                <div className={`text-[10px] mt-1 text-right font-medium opacity-70 ${isMine ? "text-cat-crust" : "text-cat-overlay1"}`}>
                    {format(new Date(message.timestamp * 1000), "h:mm a")}
                </div>
            </div>
        </div>
    );
}
