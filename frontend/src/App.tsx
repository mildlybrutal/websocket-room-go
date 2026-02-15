import { BrowserRouter, Routes, Route, Navigate } from "react-router-dom";
import { AuthProvider, useAuth } from "./context/AuthContext";
import { Login } from "./components/Auth/Login";
import { Signup } from "./components/Auth/Signup";
import { LandingPage } from "./components/LandingPage";
import { MainLayout } from "./components/Layout/MainLayout";
import { ChatArea } from "./components/Chat/ChatArea";

function HomeRoute() {
    const { isAuthenticated } = useAuth();
    return isAuthenticated ? <MainLayout /> : <LandingPage />;
}

function App() {
    return (
        <AuthProvider>
            <BrowserRouter>
                <Routes>
                    <Route path="/login" element={<Login />} />
                    <Route path="/signup" element={<Signup />} />
                    <Route path="/" element={<HomeRoute />}>
                        <Route path="room/:roomId" element={<ChatArea />} />
                        <Route index element={
                            <div className="flex-1 flex items-center justify-center bg-cat-base">
                                <div className="text-center">
                                    <h2 className="text-2xl font-bold text-cat-text mb-2">Welcome</h2>
                                    <p className="text-cat-overlay0">Select a room to start chatting</p>
                                </div>
                            </div>
                        } />
                    </Route>
                    <Route path="*" element={<Navigate to="/" replace />} />
                </Routes>
            </BrowserRouter>
        </AuthProvider>
    );
}

export default App;
