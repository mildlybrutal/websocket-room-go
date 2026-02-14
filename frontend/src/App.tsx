import { BrowserRouter, Routes, Route, Navigate } from "react-router-dom";
import { AuthProvider } from "./context/AuthContext";
import { Login } from "./components/Auth/Login";
import { Signup } from "./components/Auth/Signup";
import { MainLayout } from "./components/Layout/MainLayout";
import { ChatArea } from "./components/Chat/ChatArea";

function App() {
    return (
        <AuthProvider>
            <BrowserRouter>
                <Routes>
                    <Route path="/login" element={<Login />} />
                    <Route path="/signup" element={<Signup />} />
                    <Route path="/" element={<MainLayout />}>
                        <Route path="room/:roomId" element={<ChatArea />} />
                        <Route index element={<Navigate to="/room/general" replace />} />
                    </Route>
                    <Route path="*" element={<Navigate to="/" replace />} />
                </Routes>
            </BrowserRouter>
        </AuthProvider>
    );
}

export default App;
