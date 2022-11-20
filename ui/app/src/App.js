import './App.css';
import BrowserWaringPanel from "./components/BrowserWaringPanel";
import Link from '@mui/joy/Link';
import Typography from "@mui/joy/Typography";
import { CssVarsProvider } from '@mui/joy/styles';
import ErrorPanel from "./components/ErrorPanel";

function App() {
    let message = {
        code: 1000
    }

    try {
        const body = document.getElementById("body")
        message = JSON.parse(atob(body.dataset.message))
    } catch (e) {

    }

    return (
        <CssVarsProvider
            defaultMode="system"
            modeStorageKey="identify-system-mode"
        >
            <div className="App">
                {(message.code === 1000) ? (
                    <BrowserWaringPanel message={ message }/>
                ) : (
                    <ErrorPanel message={ message } />
                )}

                <Typography level="body3" sx={{mb: 2}}>
                    Hosted by tunl.online, view project on the <Link href="https://github.com/black40x/tunl-cli" target="_blank">GitHub</Link>
                </Typography>
            </div>
        </CssVarsProvider>
    );
}

export default App;
