import { useMemo } from 'react';
import { Link, useSearchParams } from 'react-router-dom';
import { ArrowLeft, ExternalLink, ShieldAlert, Video } from 'lucide-react';

export default function Telemedicine() {
    const [searchParams] = useSearchParams();

    const joinUrl = searchParams.get('join_url') || '';
    const title = searchParams.get('title') || 'Telemedicine Session';
    const doctor = searchParams.get('doctor') || 'Doctor';

    const iframeSrc = useMemo(() => {
        if (!joinUrl) return '';
        return joinUrl;
    }, [joinUrl]);

    return (
        <div className="min-h-screen bg-[#f6f8fa] flex flex-col font-sans">
            <header className="h-16 bg-white border-b border-gray-100 flex items-center justify-between px-6 sticky top-0 z-10">
                <div className="flex items-center gap-3 min-w-0">
                    <div className="w-9 h-9 bg-brand rounded-xl flex items-center justify-center shrink-0">
                        <Video className="h-4 w-4 text-white" />
                    </div>
                    <div className="min-w-0">
                        <h1 className="text-sm font-bold text-gray-900 truncate">{title}</h1>
                        <p className="text-xs text-gray-500 truncate">Meeting with {doctor}</p>
                    </div>
                </div>

                <div className="flex items-center gap-3">
                    <Link
                        to="/appointments"
                        className="inline-flex items-center gap-2 px-4 py-2 rounded-xl text-sm font-medium bg-gray-50 text-gray-700 hover:bg-gray-100 transition-colors"
                    >
                        <ArrowLeft className="h-4 w-4" />
                        Back
                    </Link>
                </div>
            </header>

            <main className="flex-1 p-4 md:p-6">
                <div className="max-w-[1400px] mx-auto h-[calc(100vh-6.5rem)]">
                    {!joinUrl ? (
                        <div className="h-full bg-white rounded-3xl border border-gray-100 shadow-sm flex items-center justify-center p-8 text-center">
                            <div className="max-w-md">
                                <div className="w-16 h-16 mx-auto mb-4 bg-amber-100 rounded-full flex items-center justify-center">
                                    <ShieldAlert className="h-8 w-8 text-amber-600" />
                                </div>
                                <h2 className="text-xl font-bold text-gray-900">Missing meeting link</h2>
                                <p className="mt-2 text-sm text-gray-500">
                                    The meeting URL was not provided. Go back to your appointments and open the session again.
                                </p>
                                <Link
                                    to="/appointments"
                                    className="inline-flex items-center gap-2 mt-6 px-5 py-3 rounded-xl bg-brand text-white font-semibold hover:bg-brand-dark transition-colors"
                                >
                                    <ArrowLeft className="h-4 w-4" />
                                    Return to appointments
                                </Link>
                            </div>
                        </div>
                    ) : (
                        <div className="h-full rounded-3xl overflow-hidden border border-gray-100 shadow-xl bg-white flex flex-col">
                            <div className="px-4 py-3 border-b border-gray-100 bg-gray-50 flex items-center justify-between gap-4">
                                <div className="min-w-0">
                                    <p className="text-xs font-bold uppercase tracking-wider text-gray-400">In-app meeting</p>
                                    <p className="text-sm text-gray-500 truncate">{joinUrl}</p>
                                </div>
                                <a
                                    href={joinUrl}
                                    target="_blank"
                                    rel="noreferrer"
                                    className="inline-flex items-center gap-2 px-4 py-2 rounded-xl bg-gray-900 text-white text-sm font-semibold hover:bg-brand transition-colors"
                                >
                                    <ExternalLink className="h-4 w-4" />
                                    Open fallback
                                </a>
                            </div>

                            <div className="flex-1 bg-black">
                                <iframe
                                    title="Telemedicine meeting"
                                    src={iframeSrc}
                                    className="w-full h-full border-0"
                                    allow="camera; microphone; fullscreen; display-capture; autoplay"
                                    allowFullScreen
                                />
                            </div>
                        </div>
                    )}
                </div>
            </main>
        </div>
    );
}
