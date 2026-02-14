import { Outlet, Navigate } from "react-router-dom";
import { useAuth } from "../../context/AuthContext";
import { Sidebar } from "../Chat/Sidebar";

export function MainLayout() {
    const { isAuthenticated } = useAuth();

    if (!isAuthenticated) {
        return <Navigate to="/login" replace />;
    }

    return (
        <div className="flex h-screen bg-cat-base text-cat-text overflow-hidden font-sans selection:bg-cat-surface2 selection:text-cat-text">
            <Sidebar />
            <main className="flex-1 flex flex-col relative bg-cat-base/50 backdrop-blur-sm">
                <Outlet />
            </main>
        </div>
    );
}
