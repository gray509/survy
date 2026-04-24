import { useState, useEffect } from "react";
import {Routes, Route, useNavigate, Navigate } from "react-router-dom";

export default function App() {
  const [accessToken, setAccessToken] = useState(null)
  //const navigate = useNavigate()
  useEffect(() => {
    let called = false; //dev 
    async function refreshSession() {
      if (called) return;//dev
      called = true; //dev
      try {
        const res = await fetch("/v0/refresh", {
          method: "POST",
          credentials: "include",
        });
        if (!res.ok) {
          setAccessToken("");// unauthenticated
          return;
        }
        const data = await res.json();
        setAccessToken(data.access_token); // authenticated
      } catch (err) {
        console.log("No session");
      }
    }

    refreshSession();
  }, []);
  return (
    <>
      <Routes>
        <Route path="/" element={<Home accessToken={accessToken} />} />
        <Route path="/login" element={accessToken ? <Navigate to="/" /> : <Login setAccessToken={setAccessToken}/>} />
      </Routes> 
    </>
    
  );
}

function Login({setAccessToken}){
  const [email, setEmail] = useState("")
  const [password, setPassword] = useState("")
  const navigate = useNavigate()
  
  async function handleSubmit(e) {
    e.preventDefault();
    // send to backend
    try{
      const res = await fetch("/v0/login", {
        method: "POST",
        headers: {
        'Content-Type': 'application/json',
      },
        body: JSON.stringify({ email, password }),
      });
      const data = await res.json();
      if (!res.ok) {
        throw new Error(`Failed to login: ${data.error}`);
      }
      // data do something
      const user = {
        id: data.id,
        createdAt: data.created_at,
        updatedAt: data.updated_at,
        email: data.email,
      }
      setAccessToken(data.access_token);
      sessionStorage.setItem("userInfo", JSON.stringify(user));
      navigate("/");
    } 
    catch(error){
      alert(`Error: ${error.message}`)
    }
    
    
  }

  async function handleCreateAccount(e) {
    e.preventDefault();
    // send to backend
    try{
      const res = await fetch("/v0/signup", {
        method: "POST",
        headers: {
        'Content-Type': 'application/json',
      },
        body: JSON.stringify({ email, password }),
      });
      const data = await res.json();
      if (!res.ok) {
        throw new Error(`Failed to create user: ${data.error}`);
      }
      navigate("/");

    } 
    catch(error){
      alert(`Error: ${error.message}`)
    }
  }

  return (
     <form className="floating-form" onSubmit={handleSubmit}>
      test@test.com
      pass
      <h2>Login</h2>

      <div className="input-group">
        <input
          type="email"
          required
          value={email}
          onChange={(e) => setEmail(e.target.value)}
        />
        <label>Email Address</label>
      </div>

      <div className="input-group">
        <input
          type="password"
          required
          value={password}
          onChange={(e) => setPassword(e.target.value)}
        />
        <label>Password</label>
      </div>
      <button type="submit">Sign In</button>
      <button onClick={handleCreateAccount} type="button">Create account</button>
    </form>
  )
}

function Home({accessToken}){
  if (accessToken === null) {
    return <div>Loading...</div>;
  }

  if (accessToken === "") {
    return <Navigate to="/login" />;
  }

  return <h1>Home</h1>;
}