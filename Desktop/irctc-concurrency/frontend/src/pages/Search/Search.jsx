import { useState, useEffect } from "react";
import { getAllTrains, searchTrains } from "../../api/trainApi";
import TrainCard from "../../components/TrainCard";
import Loader from "../../components/Loader";
import { MESSAGES } from "../../constants/constants";
import { useLocation } from "react-router-dom"; 

const STATIONS = [
  "Chennai",
  "Mumbai", 
  "Delhi",
  "Bangalore",
  "Kolkata",
  "Hyderabad",
  "Goa",
  "Pune",
  "Mangalore",
  "Mysore",
  "Coimbatore",
];

const Search = () => {
  const [trains, setTrains]           = useState([]);
  const [loading, setLoading]         = useState(false);
  const [error, setError]             = useState("");
  const [source, setSource]           = useState("");
  const [destination, setDestination] = useState("");
  const location = useLocation();
  const [journeyDate, setJourneyDate] = useState(
  location.state?.journeyDate || new Date().toISOString().split("T")[0]
);

  useEffect(() => {
  fetchAllTrains(journeyDate);
}, [journeyDate]);

  const fetchAllTrains = async (date) => {
  setLoading(true);
  try {
    const res = await getAllTrains(date);
    setTrains(res.data || []);
  } catch {
    setError(MESSAGES.SERVER_ERROR);
  } finally {
    setLoading(false);
  }
};

const handleSearch = async (e) => {
  e.preventDefault();
  if (!source || !destination) return;
  setLoading(true);
  setError("");
  setTrains([]);
  try {
    const res = await searchTrains(source, destination, journeyDate);
    if (res.success) {
      setTrains(res.data || []);
      if (!res.data || res.data.length === 0) setError(MESSAGES.NO_TRAINS);
    } else {
      setTrains([]);
      setError(MESSAGES.NO_TRAINS);
    }
  } catch {
    setTrains([]);
    setError(MESSAGES.SERVER_ERROR);
  } finally {
    setLoading(false);
  }
};

  const inputStyle = {
    padding: "10px 14px",
    borderRadius: "8px",
    border: "1px solid #D1D5DB",
    fontSize: "14px",
    outline: "none",
    width: "100%",
  };

  return (
    <div style={{ maxWidth: "860px", margin: "0 auto", padding: "32px 16px" }}>

      {/* Header */}
      <h1 style={{ fontSize: "24px", fontWeight: "700", color: "#111827", marginBottom: "8px" }}>
        🚂 Search Trains
      </h1>
      <p style={{ color: "#6B7280", marginBottom: "24px" }}>
        Find and book train tickets across India
      </p>

      {/* Search form */}
      <form onSubmit={handleSearch} style={{
        background: "white",
        borderRadius: "12px",
        padding: "24px",
        boxShadow: "0 1px 6px rgba(0,0,0,0.08)",
        marginBottom: "32px",
        display: "grid",
        gridTemplateColumns: "1fr 1fr 1fr auto",
        gap: "16px",
        alignItems: "end",
      }}>
        <div style={{ position: "relative" }}>
  <label style={{ fontSize: "12px", fontWeight: "600", color: "#374151", display: "block", marginBottom: "6px" }}>
    FROM
  </label>
  <input
    style={inputStyle}
    placeholder="e.g. Chennai"
    value={source}
    onChange={(e) => setSource(e.target.value)}
    autoComplete="off"
  />
  {source && STATIONS.filter(s => s.toLowerCase().startsWith(source.toLowerCase()) && s.toLowerCase() !== source.toLowerCase()).length > 0 && (
    <div style={{
      position: "absolute", top: "100%", left: 0, right: 0,
      background: "white", border: "1px solid #D1D5DB",
      borderRadius: "8px", boxShadow: "0 4px 12px rgba(0,0,0,0.1)",
      zIndex: 10, marginTop: "4px"
    }}>
      {STATIONS
        .filter(s => s.toLowerCase().startsWith(source.toLowerCase()) && s.toLowerCase() !== source.toLowerCase())
        .map(s => (
          <div
            key={s}
            onClick={() => setSource(s)}
            style={{
              padding: "10px 14px", cursor: "pointer",
              fontSize: "14px", color: "#111827",
              borderBottom: "1px solid #F3F4F6",
            }}
            onMouseEnter={e => e.target.style.background = "#F3F4F6"}
            onMouseLeave={e => e.target.style.background = "white"}
          >
            🚉 {s}
          </div>
        ))}
    </div>
  )}
</div>

<div style={{ position: "relative" }}>
  <label style={{ fontSize: "12px", fontWeight: "600", color: "#374151", display: "block", marginBottom: "6px" }}>
    TO
  </label>
  <input
    style={inputStyle}
    placeholder="e.g. Mumbai"
    value={destination}
    onChange={(e) => setDestination(e.target.value)}
    autoComplete="off"
  />
  {destination && STATIONS.filter(s => s.toLowerCase().startsWith(destination.toLowerCase()) && s.toLowerCase() !== destination.toLowerCase()).length > 0 && (
    <div style={{
      position: "absolute", top: "100%", left: 0, right: 0,
      background: "white", border: "1px solid #D1D5DB",
      borderRadius: "8px", boxShadow: "0 4px 12px rgba(0,0,0,0.1)",
      zIndex: 10, marginTop: "4px"
    }}>
      {STATIONS
        .filter(s => s.toLowerCase().startsWith(destination.toLowerCase()) && s.toLowerCase() !== destination.toLowerCase())
        .map(s => (
          <div
            key={s}
            onClick={() => setDestination(s)}
            style={{
              padding: "10px 14px", cursor: "pointer",
              fontSize: "14px", color: "#111827",
              borderBottom: "1px solid #F3F4F6",
            }}
            onMouseEnter={e => e.target.style.background = "#F3F4F6"}
            onMouseLeave={e => e.target.style.background = "white"}
          >
            🚉 {s}
          </div>
        ))}
    </div>
  )}
</div>

        <input
            type="date"
            style={inputStyle}
            value={journeyDate}
            onChange={(e) => setJourneyDate(e.target.value)}
            min={new Date().toISOString().split("T")[0]}
          />

        <button type="submit" style={{
          background: "#2563EB", color: "white",
          border: "none", borderRadius: "8px",
          padding: "10px 24px", fontWeight: "600",
          fontSize: "14px", cursor: "pointer",
          whiteSpace: "nowrap",
        }}>
          Search
        </button>
      </form>

      {/* Results */}
      {loading && <Loader message="Searching trains..." />}
      {error && (
        <div style={{
          background: "#FEF2F2", border: "1px solid #FECACA",
          borderRadius: "8px", padding: "16px",
          color: "#DC2626", fontSize: "14px", marginBottom: "16px"
        }}>
          {error}
        </div>
      )}
      {!loading && (
        <div style={{ display: "flex", flexDirection: "column", gap: "12px" }}>
          {trains.map((train) => (
            <TrainCard
              key={train.id}
              train={train}
              journeyDate={journeyDate || new Date().toISOString().split("T")[0]}
            />
          ))}
        </div>
      )}
    </div>
  );
};

export default Search;