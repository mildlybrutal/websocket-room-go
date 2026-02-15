import { useState } from "react";
import { Link, useNavigate } from "react-router-dom";
import { useAuth } from "../../context/AuthContext";
import { api } from "../../services/api";
import { Lock, Mail, MessageCircle } from "lucide-react";

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
            login(data.token, { id: data.id, username: data.username });
            navigate("/");
        } catch (err: any) {
            setError(err.message);
        }
    };

    return (
        <div className="relative flex items-center justify-center min-h-screen bg-cat-crust text-cat-text overflow-hidden">
            {/* Animated background blobs */}
            <div className="absolute inset-0 overflow-hidden pointer-events-none">
                <div className="absolute -top-40 -left-40 w-96 h-96 bg-cat-mauve/10 rounded-full blur-3xl animate-blob" />
                <div className="absolute -bottom-40 -right-40 w-96 h-96 bg-cat-blue/10 rounded-full blur-3xl animate-blob animation-delay-2000" />
                <div className="absolute top-1/2 left-1/2 -translate-x-1/2 -translate-y-1/2 w-80 h-80 bg-cat-pink/8 rounded-full blur-3xl animate-blob animation-delay-4000" />
            </div>

            <div className="relative z-10 w-full max-w-md p-8 space-y-8 bg-cat-mantle/80 backdrop-blur-xl rounded-2xl shadow-2xl shadow-cat-crust/50 border border-cat-surface0/50">
                <div className="text-center">
                    <div className="inline-flex items-center justify-center w-14 h-14 rounded-2xl bg-cat-mauve/15 mb-4">
                        <MessageCircle className="text-cat-mauve" size={28} />
                    </div>
                    <h2 className="text-3xl font-extrabold tracking-tight text-cat-text">Welcome Back</h2>
                    <p className="mt-2 text-sm text-cat-overlay0">Sign in to your account</p>
                </div>
                <form className="mt-8 space-y-6" onSubmit={handleSubmit}>
                    {error && <div className="p-3 text-sm text-cat-red bg-cat-red/10 rounded-lg border border-cat-red/20">{error}</div>}
                    <div className="space-y-4">
                        <div className="relative group">
                            <Mail className="absolute left-3 top-3 text-cat-overlay0 transition-colors group-focus-within:text-cat-mauve" size={20} />
                            <input
                                type="email"
                                required
                                className="w-full pl-10 pr-4 py-2.5 bg-cat-base/70 border border-cat-surface0 rounded-xl focus:ring-2 focus:ring-cat-mauve/50 focus:border-cat-mauve focus:outline-none text-cat-text placeholder-cat-overlay0 transition-all duration-200"
                                placeholder="Email address"
                                value={email}
                                onChange={(e) => setEmail(e.target.value)}
                            />
                        </div>
                        <div className="relative group">
                            <Lock className="absolute left-3 top-3 text-cat-overlay0 transition-colors group-focus-within:text-cat-mauve" size={20} />
                            <input
                                type="password"
                                required
                                className="w-full pl-10 pr-4 py-2.5 bg-cat-base/70 border border-cat-surface0 rounded-xl focus:ring-2 focus:ring-cat-mauve/50 focus:border-cat-mauve focus:outline-none text-cat-text placeholder-cat-overlay0 transition-all duration-200"
                                placeholder="Password"
                                value={password}
                                onChange={(e) => setPassword(e.target.value)}
                            />
                        </div>
                    </div>

                    <button
                        type="submit"
                        className="w-full py-2.5 px-4 bg-cat-mauve hover:bg-cat-mauve/80 text-cat-crust font-semibold rounded-xl transition-all duration-200 transform hover:scale-[1.02] hover:shadow-lg hover:shadow-cat-mauve/25 active:scale-[0.98]"
                    >
                        Sign in
                    </button>
                </form>
                <div className="text-center text-sm text-cat-overlay0">
                    Don't have an account?{" "}
                    <Link to="/signup" className="font-medium text-cat-mauve hover:text-cat-lavender transition-colors">
                        Sign up
                    </Link>
                </div>
            </div>
        </div>
    );
}
