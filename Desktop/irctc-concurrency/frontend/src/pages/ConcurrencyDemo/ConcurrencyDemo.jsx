import { useState, useEffect } from "react";
import { createBooking, checkSeatStatus } from "../../api/bookingApi";
import { getAllTrains, getSeats } from "../../api/trainApi";

const BATCH_SIZE = 100;

const ConcurrencyDemo = () => {
  const [results, setResults] = useState([]);
  const [running, setRunning] = useState(false);

  const [trains, setTrains] = useState([]);
  const [seats, setSeats] = useState([]);

  const [trainId, setTrainId] = useState("");
  const [seatId, setSeatId] = useState("");
  const [userCount, setUserCount] = useState(10);

  const booked = results.filter((r) =>
    r.status.includes("BOOKED")
  ).length;

  const rejected = results.filter((r) =>
    r.status.includes("REJECTED")
  ).length;

  useEffect(() => {
    fetchTrains();
  }, []);

  useEffect(() => {
    if (trainId) {
      fetchSeats();
    }
  }, [trainId]);

  const fetchTrains = async () => {
    try {
      const journeyDate = new Date().toISOString().split("T")[0]
      const res = await getAllTrains(journeyDate)


      if (res.success) {
        setTrains(res.data || []);

        if (res.data?.length > 0) {
          setTrainId(res.data[0].id);
        }
      }
    } catch (err) {
      console.error("Failed to fetch trains", err);
    }
  };

  const fetchSeats = async () => {
    try {
      const journeyDate = new Date().toISOString().split("T")[0]
      const res = await getSeats(trainId, journeyDate)


      if (res.success) {
        setSeats(res.data || []);

        if (res.data?.length > 0) {
          setSeatId(res.data[0].id);
        }
      }
    } catch (err) {
      console.error("Failed to fetch seats", err);
    }
  };


  const runDemo = async () => {
    setRunning(true);
    setResults([]);

    const journeyDate = new Date()
      .toISOString()
      .split("T")[0];

    try {
      const statusRes = await checkSeatStatus(
        trainId,
        seatId,
        journeyDate
      );

      if (
        statusRes.success &&
        !statusRes.data.is_available
      ) {
        alert(
          " Seat already booked! Choose another seat."
        );

        setRunning(false);
        return;
      }
    } catch {
     
    }

    const allResults = [];

    const totalBatches = Math.ceil(
      userCount / BATCH_SIZE
    );

    for (let batch = 0; batch < totalBatches; batch++) {
      const start = batch * BATCH_SIZE;

      const end = Math.min(
        start + BATCH_SIZE,
        userCount
      );

      const promises = Array.from(
        { length: end - start },
        (_, i) => {
          const globalIndex = start + i;

          return createBooking({
            user_id: (globalIndex % 3) + 1,
            train_id: trainId,
            seat_id: seatId,
            journey_date: journeyDate,
          })
            .then((res) => ({
              user: `User ${(globalIndex % 3) + 1}`,
              request: globalIndex + 1,
              status: " BOOKED",
              message: res.message,
              color: "#059669",
            }))
            .catch((err) => ({
              user: `User ${(globalIndex % 3) + 1}`,
              request: globalIndex + 1,
              status: " REJECTED",
              message:
                err.response?.data?.message ||
                "Seat already taken",
              color: "#EF4444",
            }));
        }
      );

      const batchResults = await Promise.all(promises);

      allResults.push(...batchResults);

      setResults([...allResults]);
    }

    setRunning(false);
  };

  return (
    <div
      style={{
        maxWidth: "760px",
        margin: "0 auto",
        padding: "32px 16px",
      }}
    >
      {/* Header */}
      <div
        style={{
          background:
            "linear-gradient(135deg, #1E3A5F, #2563EB)",
          borderRadius: "16px",
          padding: "28px 32px",
          color: "white",
          marginBottom: "28px",
        }}
      >
        <h1
          style={{
            fontSize: "24px",
            fontWeight: "700",
            margin: "0 0 8px",
          }}
        >
          ⚡ Concurrency Demo
        </h1>

        <p
          style={{
            opacity: 0.85,
            margin: 0,
            fontSize: "14px",
          }}
        >
          Fire multiple booking requests simultaneously
          for the same seat. Watch how Redis ensures only
          ONE succeeds.
        </p>
      </div>

      {/* Controls */}
      <div
        style={{
          background: "white",
          borderRadius: "12px",
          padding: "24px",
          boxShadow: "0 1px 6px rgba(0,0,0,0.08)",
          marginBottom: "24px",
          display: "grid",
          gridTemplateColumns: "1fr 1fr 1fr auto",
          gap: "16px",
          alignItems: "end",
        }}
      >
        {/* Train Selector */}
        <div>
          <label
            style={{
              fontSize: "12px",
              fontWeight: "600",
              color: "#374151",
              display: "block",
              marginBottom: "6px",
            }}
          >
            TRAIN
          </label>

          <select
            value={trainId}
            onChange={(e) => {
              setTrainId(Number(e.target.value));
              setSeatId("");
            }}
            style={{
              width: "100%",
              padding: "10px 14px",
              borderRadius: "8px",
              border: "1px solid #D1D5DB",
              fontSize: "14px",
              background: "white",
            }}
          >
            {trains.map((t) => (
              <option key={t.id} value={t.id}>
                {t.train_name} ({t.source} →{" "}
                {t.destination})
              </option>
            ))}
          </select>
        </div>

        {/* Seat Selector */}
        <div>
          <label
            style={{
              fontSize: "12px",
              fontWeight: "600",
              color: "#374151",
              display: "block",
              marginBottom: "6px",
            }}
          >
            SEAT
          </label>

          <select
            value={seatId}
            onChange={(e) =>
              setSeatId(Number(e.target.value))
            }
            style={{
              width: "100%",
              padding: "10px 14px",
              borderRadius: "8px",
              border: "1px solid #D1D5DB",
              fontSize: "14px",
              background: "white",
            }}
          >
            {seats.map((s) => (
              <option key={s.id} value={s.id}>
                {s.seat_number} —{" "}
                {s.seat_type.replace("_", " ")}{" "}
                {s.is_available
                  ? " Available"
                  : " Taken"}
              </option>
            ))}
          </select>
        </div>

        {/* User Count */}
        <div>
          <label
            style={{
              fontSize: "12px",
              fontWeight: "600",
              color: "#374151",
              display: "block",
              marginBottom: "6px",
            }}
          >
            SIMULTANEOUS USERS
          </label>

          <input
            type="number"
            min="2"
            value={userCount}
            onChange={(e) =>
              setUserCount(Number(e.target.value))
            }
            style={{
              width: "100%",
              padding: "10px 14px",
              borderRadius: "8px",
              border: "1px solid #D1D5DB",
              fontSize: "14px",
            }}
          />
        </div>

        {/* Fire Button */}
        <button
          onClick={runDemo}
          disabled={running || !seatId}
          style={{
            background: running
              ? "#9CA3AF"
              : "#EF4444",
            color: "white",
            border: "none",
            borderRadius: "8px",
            padding: "10px 24px",
            fontWeight: "700",
            fontSize: "14px",
            cursor: running
              ? "not-allowed"
              : "pointer",
            whiteSpace: "nowrap",
          }}
        >
          {running
            ? `Firing... ${results.length}/${userCount}`
            : "🚀 Fire!"}
        </button>
      </div>

      {/* Score Cards */}
      {results.length > 0 && (
        <div
          style={{
            display: "grid",
            gridTemplateColumns: "1fr 1fr 1fr",
            gap: "16px",
            marginBottom: "20px",
          }}
        >
          {[
            {
              label: "Total Fired",
              value: results.length,
              color: "#2563EB",
              bg: "#EFF6FF",
            },
            {
              label: " Booked",
              value: booked,
              color: "#059669",
              bg: "#D1FAE5",
            },
            {
              label: " Rejected",
              value: rejected,
              color: "#EF4444",
              bg: "#FEE2E2",
            },
          ].map((s) => (
            <div
              key={s.label}
              style={{
                background: s.bg,
                borderRadius: "10px",
                padding: "16px",
                textAlign: "center",
              }}
            >
              <p
                style={{
                  fontSize: "32px",
                  fontWeight: "700",
                  color: s.color,
                  margin: "0 0 4px",
                }}
              >
                {s.value}
              </p>

              <p
                style={{
                  fontSize: "13px",
                  color: s.color,
                  margin: 0,
                  fontWeight: "500",
                }}
              >
                {s.label}
              </p>
            </div>
          ))}
        </div>
      )}

      {/* Results */}
      {results.length > 0 && (
        <div
          style={{
            background: "white",
            borderRadius: "12px",
            padding: "20px",
            boxShadow:
              "0 1px 6px rgba(0,0,0,0.08)",
          }}
        >
          <h3
            style={{
              fontSize: "14px",
              fontWeight: "700",
              marginBottom: "16px",
              color: "#111827",
            }}
          >
            Results — {results.length} requests fired
          </h3>

          <div
            style={{
              display: "flex",
              flexDirection: "column",
              gap: "8px",
            }}
          >
            {results.slice(0, 200).map((r, i) => (
              <div
                key={i}
                style={{
                  display: "flex",
                  justifyContent:
                    "space-between",
                  alignItems: "center",
                  padding: "10px 14px",
                  borderRadius: "8px",
                  background: `${r.color}10`,
                  border: `1px solid ${r.color}30`,
                }}
              >
                <span
                  style={{
                    fontSize: "13px",
                    color: "#374151",
                  }}
                >
                  Request #{r.request} · {r.user}
                </span>

                <span
                  style={{
                    fontSize: "12px",
                    fontWeight: "700",
                    color: r.color,
                  }}
                >
                  {r.status}
                </span>
              </div>
            ))}
          </div>

          {/* Success */}
          {!running && booked === 1 && (
            <div
              style={{
                marginTop: "16px",
                padding: "14px",
                background: "#D1FAE5",
                borderRadius: "8px",
                border: "1px solid #6EE7B7",
                textAlign: "center",
              }}
            >
              <p
                style={{
                  color: "#065F46",
                  fontWeight: "700",
                  margin: 0,
                }}
              >
                ✅ Exactly 1 booking succeeded.
                Redis locking worked perfectly!
              </p>
            </div>
          )}

          {/* Failure */}
          {!running && booked > 1 && (
            <div
              style={{
                marginTop: "16px",
                padding: "14px",
                background: "#FEE2E2",
                borderRadius: "8px",
                border: "1px solid #FCA5A5",
                textAlign: "center",
              }}
            >
              <p
                style={{
                  color: "#991B1B",
                  fontWeight: "700",
                  margin: 0,
                }}
              >
                ⚠️ Double booking detected!{" "}
                {booked} bookings succeeded.
              </p>
            </div>
          )}

          {/* Running */}
          {running && (
            <div
              style={{
                marginTop: "16px",
                padding: "14px",
                background: "#EFF6FF",
                borderRadius: "8px",
                border: "1px solid #BFDBFE",
                textAlign: "center",
              }}
            >
              <p
                style={{
                  color: "#1E3A5F",
                  fontWeight: "600",
                  margin: 0,
                }}
              >
                ⏳ Firing requests...{" "}
                {results.length} / {userCount}
              </p>
            </div>
          )}
        </div>
      )}
    </div>
  );
};

export default ConcurrencyDemo;