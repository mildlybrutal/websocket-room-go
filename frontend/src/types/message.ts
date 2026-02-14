export interface Message {
    type: "message" | "join_room" | "leave_room";
    content: string;
    room?: string;
    sender?: string;
    data?: Date;
    timestamp?: string;
}
