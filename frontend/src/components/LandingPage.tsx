import { Link } from "react-router-dom";
import { MessageCircle, Users, Lock, Zap, Globe, Shield, Code, Server } from "lucide-react";

export function LandingPage() {
    return (
        <div className="relative min-h-screen bg-cat-crust text-cat-text overflow-hidden">
            {/* Animated background blobs */}
            <div className="absolute inset-0 overflow-hidden pointer-events-none">
                <div className="absolute -top-60 -left-60 w-[500px] h-[500px] bg-cat-mauve/8 rounded-full blur-3xl animate-blob" />
                <div className="absolute top-1/4 -right-40 w-[450px] h-[450px] bg-cat-blue/8 rounded-full blur-3xl animate-blob animation-delay-2000" />
                <div className="absolute -bottom-40 left-1/3 w-[400px] h-[400px] bg-cat-pink/6 rounded-full blur-3xl animate-blob animation-delay-4000" />
                <div className="absolute top-2/3 right-1/4 w-[350px] h-[350px] bg-cat-teal/5 rounded-full blur-3xl animate-blob animation-delay-6000" />
            </div>

            {/* Navigation */}
            <nav className="relative z-10 flex items-center justify-between px-6 md:px-12 py-5">
                <div className="flex items-center gap-3">
                    <div className="p-2 rounded-xl bg-cat-mauve/15">
                        <MessageCircle className="text-cat-mauve" size={24} />
                    </div>
                    <span className="text-xl font-bold text-cat-text tracking-tight">RoomChat</span>
                </div>
                <div className="flex items-center gap-3">
                    <Link
                        to="/login"
                        className="px-5 py-2 text-sm font-medium text-cat-subtext1 hover:text-cat-text rounded-lg transition-colors duration-200"
                    >
                        Sign In
                    </Link>
                    <Link
                        to="/signup"
                        className="px-5 py-2 text-sm font-semibold bg-cat-mauve hover:bg-cat-mauve/80 text-cat-crust rounded-xl transition-all duration-200 hover:shadow-lg hover:shadow-cat-mauve/25"
                    >
                        Get Started
                    </Link>
                </div>
            </nav>

            {/* Hero Section */}
            <section className="relative z-10 flex flex-col items-center text-center px-6 pt-20 pb-24 md:pt-32 md:pb-32">
                <div className="inline-flex items-center gap-2 px-4 py-1.5 rounded-full bg-cat-surface0/50 border border-cat-surface1/50 text-cat-subtext0 text-xs font-medium mb-8 backdrop-blur-sm">
                    <Zap size={14} className="text-cat-yellow" />
                    Real-time WebSocket powered
                </div>
                <h1 className="text-5xl md:text-7xl font-extrabold tracking-tight leading-tight max-w-4xl">
                    <span className="text-cat-text">Chat rooms,</span>
                    <br />
                    <span className="bg-gradient-to-r from-cat-mauve via-cat-pink to-cat-blue bg-clip-text text-transparent">reimagined.</span>
                </h1>
                <p className="mt-6 text-lg md:text-xl text-cat-overlay1 max-w-2xl leading-relaxed">
                    A blazing-fast, real-time chat platform built with Go and WebSockets.
                    Create rooms, invite friends, and start conversations that matter.
                </p>
                <div className="flex flex-col sm:flex-row items-center gap-4 mt-10">
                    <Link
                        to="/signup"
                        className="px-8 py-3 text-base font-semibold bg-cat-mauve hover:bg-cat-mauve/80 text-cat-crust rounded-xl transition-all duration-200 transform hover:scale-[1.03] hover:shadow-xl hover:shadow-cat-mauve/25 active:scale-[0.98]"
                    >
                        Start Chatting — It's Free
                    </Link>
                    <Link
                        to="/login"
                        className="px-8 py-3 text-base font-medium text-cat-subtext1 hover:text-cat-text bg-cat-surface0/40 hover:bg-cat-surface0/60 border border-cat-surface1/50 rounded-xl transition-all duration-200 backdrop-blur-sm"
                    >
                        I have an account
                    </Link>
                </div>
            </section>

            {/* Features Section */}
            <section className="relative z-10 px-6 md:px-12 pb-24">
                <div className="max-w-6xl mx-auto">
                    <div className="text-center mb-16">
                        <h2 className="text-3xl md:text-4xl font-bold text-cat-text">Everything you need</h2>
                        <p className="mt-3 text-cat-overlay0 text-lg">Simple, fast, and built for conversations.</p>
                    </div>
                    <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
                        {/* Feature Card 1 */}
                        <div className="group p-6 bg-cat-mantle/60 backdrop-blur-xl border border-cat-surface0/40 rounded-2xl transition-all duration-300 hover:border-cat-mauve/30 hover:shadow-lg hover:shadow-cat-mauve/5 hover:-translate-y-1">
                            <div className="inline-flex items-center justify-center w-12 h-12 rounded-xl bg-cat-blue/15 mb-5 group-hover:bg-cat-blue/25 transition-colors">
                                <Globe className="text-cat-blue" size={24} />
                            </div>
                            <h3 className="text-lg font-bold text-cat-text mb-2">Real-Time Messaging</h3>
                            <p className="text-sm text-cat-overlay0 leading-relaxed">
                                Instant message delivery powered by WebSocket connections. No delays, no polling — pure real-time.
                            </p>
                        </div>
                        {/* Feature Card 2 */}
                        <div className="group p-6 bg-cat-mantle/60 backdrop-blur-xl border border-cat-surface0/40 rounded-2xl transition-all duration-300 hover:border-cat-pink/30 hover:shadow-lg hover:shadow-cat-pink/5 hover:-translate-y-1">
                            <div className="inline-flex items-center justify-center w-12 h-12 rounded-xl bg-cat-pink/15 mb-5 group-hover:bg-cat-pink/25 transition-colors">
                                <Users className="text-cat-pink" size={24} />
                            </div>
                            <h3 className="text-lg font-bold text-cat-text mb-2">Room Management</h3>
                            <p className="text-sm text-cat-overlay0 leading-relaxed">
                                Create public or private rooms. Join existing conversations. Manage your spaces effortlessly.
                            </p>
                        </div>
                        {/* Feature Card 3 */}
                        <div className="group p-6 bg-cat-mantle/60 backdrop-blur-xl border border-cat-surface0/40 rounded-2xl transition-all duration-300 hover:border-cat-green/30 hover:shadow-lg hover:shadow-cat-green/5 hover:-translate-y-1">
                            <div className="inline-flex items-center justify-center w-12 h-12 rounded-xl bg-cat-green/15 mb-5 group-hover:bg-cat-green/25 transition-colors">
                                <Shield className="text-cat-green" size={24} />
                            </div>
                            <h3 className="text-lg font-bold text-cat-text mb-2">Secure by Default</h3>
                            <p className="text-sm text-cat-overlay0 leading-relaxed">
                                JWT authentication, encrypted connections, and proper session management keep your data safe.
                            </p>
                        </div>
                    </div>
                </div>
            </section>

            {/* Tech Stack Section */}
            <section className="relative z-10 px-6 md:px-12 pb-24">
                <div className="max-w-4xl mx-auto">
                    <div className="text-center mb-12">
                        <h2 className="text-3xl md:text-4xl font-bold text-cat-text">Built with modern tech</h2>
                        <p className="mt-3 text-cat-overlay0 text-lg">A robust stack for a reliable experience.</p>
                    </div>
                    <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
                        <div className="flex flex-col items-center gap-3 p-5 bg-cat-mantle/40 border border-cat-surface0/30 rounded-xl backdrop-blur-sm hover:border-cat-surface1 transition-colors">
                            <Server className="text-cat-sky" size={28} />
                            <div className="text-center">
                                <p className="text-sm font-bold text-cat-text">Go</p>
                                <p className="text-xs text-cat-overlay0">Backend</p>
                            </div>
                        </div>
                        <div className="flex flex-col items-center gap-3 p-5 bg-cat-mantle/40 border border-cat-surface0/30 rounded-xl backdrop-blur-sm hover:border-cat-surface1 transition-colors">
                            <Zap className="text-cat-yellow" size={28} />
                            <div className="text-center">
                                <p className="text-sm font-bold text-cat-text">WebSocket</p>
                                <p className="text-xs text-cat-overlay0">Real-time</p>
                            </div>
                        </div>
                        <div className="flex flex-col items-center gap-3 p-5 bg-cat-mantle/40 border border-cat-surface0/30 rounded-xl backdrop-blur-sm hover:border-cat-surface1 transition-colors">
                            <Code className="text-cat-sapphire" size={28} />
                            <div className="text-center">
                                <p className="text-sm font-bold text-cat-text">React</p>
                                <p className="text-xs text-cat-overlay0">Frontend</p>
                            </div>
                        </div>
                        <div className="flex flex-col items-center gap-3 p-5 bg-cat-mantle/40 border border-cat-surface0/30 rounded-xl backdrop-blur-sm hover:border-cat-surface1 transition-colors">
                            <Lock className="text-cat-mauve" size={28} />
                            <div className="text-center">
                                <p className="text-sm font-bold text-cat-text">JWT Auth</p>
                                <p className="text-xs text-cat-overlay0">Security</p>
                            </div>
                        </div>
                    </div>
                </div>
            </section>

            {/* Footer */}
            <footer className="relative z-10 border-t border-cat-surface0/30 px-6 py-8">
                <div className="max-w-6xl mx-auto flex flex-col md:flex-row items-center justify-between gap-4">
                    <div className="flex items-center gap-2 text-cat-overlay0 text-sm">
                        <MessageCircle size={16} className="text-cat-mauve" />
                        <span>RoomChat</span>
                    </div>
                    <p className="text-xs text-cat-overlay0">
                        Built with Go, WebSockets & React. Open source.
                    </p>
                </div>
            </footer>
        </div>
    );
}
