import Execution from "./pages/Execution";
import { Route, BrowserRouter as Router, Routes } from "react-router-dom";
import Navbar from "./components/Navbar";
import Login from "./pages/Login";
import Visualizador from "./pages/Visualizador";

function App() {
    return (
        <Router>
            <Navbar />
            <Routes>
                <Route path="/execution" element={<Execution />} />
                <Route path="/login" element={<Login/>} />
                <Route path="/visualizador" element={<Visualizador />} />
            </Routes>
        </Router>
    );
}

export default App;
