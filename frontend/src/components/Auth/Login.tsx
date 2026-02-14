import { useState } from "react";
import { Link, useNavigate } from "react-router-dom";
import { useAuth } from "../../context/AuthContext";
import { api } from "../../services/api";
import { Lock, Mail } from "lucide-react";

export function Login() {
    const [email, setEmail] = useState("");
    const [password, setPassword] = useState("");
    const [error, setError] = useState("");
    const { login } = useAuth();
    const navigate = useNavigate();

    const handleSubmit = async (e: React.FormEvent) => {
        e.preventDefault();
        try {
            const data = await api.login(email, password);
            login(data.token, { id: data.id, username: data.username }); // Assuming API returns this structure
            navigate("/");
        } catch (err: any) {
            setError(err.message);
        }
    };

    return (
        <div className="flex items-center justify-center min-h-screen bg-gradient-to-br from-gray-900 to-black text-white">
            <div className="w-full max-w-md p-8 space-y-8 bg-white/10 backdrop-blur-md rounded-2xl shadow-xl border border-white/20">
                <div className="text-center">
                    <h2 className="text-3xl font-extrabold tracking-tight">Welcome Back</h2>
                    <p className="mt-2 text-sm text-gray-400">Sign in to your account</p>
                </div>
                <form className="mt-8 space-y-6" onSubmit={handleSubmit}>
                    {error && <div className="p-3 text-sm text-red-500 bg-red-500/10 rounded-lg">{error}</div>}
                    <div className="space-y-4">
                        <div className="relative">
                            <Mail className="absolute left-3 top-3 text-gray-400" size={20} />
                            <input
                                type="email"
                                required
                                className="w-full pl-10 pr-4 py-2 bg-black/20 border border-white/10 rounded-lg focus:ring-2 focus:ring-blue-500 focus:outline-none text-white placeholder-gray-500"
                                placeholder="Email address"
                                value={email}
                                onChange={(e) => setEmail(e.target.value)}
                            />
                        </div>
                        <div className="relative">
                            <Lock className="absolute left-3 top-3 text-gray-400" size={20} />
                            <input
                                type="password"
                                required
                                className="w-full pl-10 pr-4 py-2 bg-black/20 border border-white/10 rounded-lg focus:ring-2 focus:ring-blue-500 focus:outline-none text-white placeholder-gray-500"
                                placeholder="Password"
                                value={password}
                                onChange={(e) => setPassword(e.target.value)}
                            />
                        </div>
                    </div>

                    <button
                        type="submit"
                        className="w-full py-2 px-4 bg-gradient-to-r from-blue-600 to-cyan-600 hover:from-blue-700 hover:to-cyan-700 text-white font-medium rounded-lg transition duration-200 transform hover:scale-[1.02]"
                    >
                        Sign in
                    </button>
                </form>
                <div className="text-center text-sm text-gray-400">
                    Don't have an account?{" "}
                    <Link to="/signup" className="font-medium text-blue-400 hover:text-blue-300">
                        Sign up
                    </Link>
                </div>
            </div>
        </div>
    );
}
