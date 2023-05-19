import "react-toastify/dist/ReactToastify.css";
import "../styles/globals.scss";
import type { AppType } from "next/dist/shared/lib/utils";
import dynamic from "next/dynamic";
const LayoutDefault = dynamic(
    () => import('../components/LayoutDefault'),
    { ssr: false }
)

import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { SSRProvider } from "react-bootstrap";

const queryClient = new QueryClient({
    defaultOptions: {
        queries: {
            // retryDelay: attemptIndex => Math.min(1000 * 2 ** attemptIndex, 30000),
            retry: false,
            staleTime: 10000, // milliSeconds
            cacheTime: 60000, // milliSeconds
            refetchOnWindowFocus: false,
        },
    },
});

const MyApp: AppType = ({ Component, pageProps }) => {
    return (
        <SSRProvider>
            <QueryClientProvider client={queryClient}>
                <LayoutDefault>
                    <Component {...pageProps} />
                </LayoutDefault>
            </QueryClientProvider>
        </SSRProvider>
    );
};

export default MyApp;
