import { BrowserRouter, Routes, Route, Navigate } from "react-router-dom";
import Navbar from "./components/Navbar";
import Search from "./pages/Search/Search";
import Booking from "./pages/Booking/Booking";
import Payment from "./pages/Payment/Payment";
import Confirmation from "./pages/Confirmation/Confirmation";
import ConcurrencyDemo from "./pages/ConcurrencyDemo/ConcurrencyDemo";
import Login from "./pages/Login/Login";
import Register from "./pages/Register/Register";
import BookingHistory from "./pages/BookingHistory/BookingHistory";
import { useAuth } from "./context/AuthContext";

const ProtectedRoute = ({ children }) => {
  const { token } = useAuth();
  if (!token) return <Navigate to="/login" />;
  return children;
};

const App = () => {
  return (
    <BrowserRouter>
      <div style={{ minHeight: "100vh", background: "#F3F4F6", fontFamily: "system-ui, sans-serif" }}>
        <Navbar />
        <Routes>
          <Route path="/"            element={<Navigate to="/login" />} />
          <Route path="/login"       element={<Login />} />
          <Route path="/register"    element={<Register />} />
          <Route path="/search"      element={<ProtectedRoute><Search /></ProtectedRoute>} />
          <Route path="/booking"     element={<ProtectedRoute><Booking /></ProtectedRoute>} />
          <Route path="/payment"     element={<ProtectedRoute><Payment /></ProtectedRoute>} />
          <Route path="/confirmation" element={<ProtectedRoute><Confirmation /></ProtectedRoute>} />
          <Route path="/demo"        element={<ProtectedRoute><ConcurrencyDemo /></ProtectedRoute>} />
          <Route path="/history"     element={<ProtectedRoute><BookingHistory /></ProtectedRoute>} />
        </Routes>
      </div>
    </BrowserRouter>
  );
};

export default App;