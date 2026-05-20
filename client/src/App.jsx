import { useState, useEffect, createContext, useContext } from "react";
import { Routes, Route, useNavigate, Navigate } from "react-router-dom";

const SurveyContext = createContext();

export default function App() {
  const [accessToken, setAccessToken] = useState(null)
  //const navigate = useNavigate()
  useEffect(() => {

    async function refreshSession() {
      try {
        const res = await fetch("/v0/refresh", {
          method: "POST",
          credentials: "include",
        });
        if (!res.ok) {
          const error = await res.json() 
          console.log(error.error)
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
        <Route path="/login" element={accessToken ? <Navigate to="/" /> : <Login setAccessToken={setAccessToken} />} />
        <Route path="/survey" element={<NewSurvey />} />
        <Route path="/survey/:id" element={<Survey accessToken={accessToken} />} />
      </Routes>
    </>

  );
}

function Login({ setAccessToken }) {
  const [email, setEmail] = useState("")
  const [password, setPassword] = useState("")
  const navigate = useNavigate()

  async function handleSubmit(e) {
    e.preventDefault();
    // send to backend
    try {
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
    catch (error) {
      alert(`Error: ${error.message}`)
    }


  }

  async function handleCreateAccount(e) {
    e.preventDefault();
    // send to backend
    try {
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
    catch (error) {
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

function Home({ accessToken }) {
  const [surveys, setSurveys] = useState([]);
  const [err, setError] = useState(null);
  const navigate = useNavigate();

  useEffect(() => {
    if (!accessToken) return;

    async function getUserSurveys() {
      try {
        const res = await fetch("/v0/survey?sort=asc", {
          headers: {
            Authorization: "Bearer " + accessToken,
          },
        });

        const data = await res.json();

        if (!res.ok) {
          setError(data.error);
          return;
        }

        setSurveys(data);
      } catch (error) {
        console.log("no data");
      }
    }

    getUserSurveys();
  }, [accessToken]);

  if (accessToken === null) {
    return <div>Loading...</div>;
  }

  if (accessToken === "") {
    return <Navigate to="/login" />;
  }

  if (err) {
    return <h1>Failed to load surveys: {err}</h1>;
  }

  return (
    <>
      <h1>My Surveys</h1>
      <button onClick={() => navigate("/survey")}>
        + Add Survey
      </button>
      <ul>
        {surveys.map((s) => (
          <li key={s.Id} onClick={() => navigate(`/survey/${s.Id}`)}>
            {s.Title} {s.updatedAt}
          </li>
        ))}
      </ul>
    </>
  );
}

function Survey({ accessToken }) {
  const { id } = useParams();
  const [err, setError] = useState(null);
  const [survey, setSurvey] = useState([]);

  useEffect(() => {
    async function getSurvey() {
      try {
        const res = await fetch(`/v0/survey/${id}`, {
          method: "GET",
          headers: {
            Authorization: "Bearer " + accessToken,
          },
        });
        const data = await res.json()
        if (!res.ok) {
          setError(data.error)
        }
        setSurvey(data)
      } catch (error) {
        console.log("no data");
      }
    }
    getSurvey();
  }, [id, accessToken])

  if (!accessToken){
     return <Navigate to="/login" />;
  }

  if (err) {
    return <h1>Failed to load survey: {err}</h1>;
  }

  return (
    <>
      <h1>survey {survey.Title}</h1>
    </>
  )
}

function NewSurvey({ accessToken }) {
  const [err, setError] = useState(null)
  const [survey, setSurvey] = useState({
    title: "",
    expiration_time: null,
    identified: true,
    max_response: 30,
    questions: [
      {
        title: "",
        types: "",
        required: true,
        options: {
          answers: []
        }
      }],
  })
  
  if (accessToken === "") {
    return <Navigate to="/login" />;
  }

  const handleSubmitSurvey = (e) => {
    async function CreateSurvey(e) {
      try {
        const res = await fetch(`/v0/survey/`, {
          method: "POST",
          headers: {
            'Authorization': "Bearer " + accessToken,
            'Content-Type': 'application/json',
          },
          body: JSON.stringify(survey)
        });
        const data = await res.json()
        if (!res.ok) {
          setError(data.error)
        }
      } catch (error) {
        console.log(err);
      }
    }
    CreateSurvey(e);
  }

  const addQuestion = (e) => {
    e.preventDefault();
    const type = e.target.value;
    const newQuestion = {
      title: "",
      type: type,
      required: true,
      options: { answers: [] }
    };

    setSurvey(prev => ({
      ...prev,
      questions: [...prev.questions, newQuestion]
    }));
  };

 
  return (
    <>
      <SurveyContext.Provider value={{ survey, setSurvey }}>
        <div className="survey-creation">
          <form id="survey">
            <Title />
            <div className="add-question">
              <button type="button" onClick={addQuestion} value="radio">Single</button>
              <button type="button" onClick={addQuestion} value="checkbox">Multiple</button>
              <button type="button" onClick={addQuestion} value="rating">Rating</button>
              <button type="button" onClick={addQuestion} value="yes/no">Yes/No -- True/Flase</button>
              <button type="button" onClick={addQuestion} value="ranking">Ranking</button>
              <button type="button" onClick={addQuestion} value="open">Open Text</button>
            </div>
            <div className="questions">
              {survey.questions.map((q, index) => {
                switch (q.type) {
                  case "radio", "checkbox", "ranking":
                    return <SingleMutipleRanking key={index} />;
                  case "rating":
                    return <Rating key={index} />;
                  case "yes/no":
                    return <Bool key={index} />;
                  case "open":
                    return <Open key={index} />;
                  default:
                    return null;
                }
              })}
            </div>
            <div className="submit-survey">
              <button type="submit" onClick={handleSubmitSurvey}>Sumbit</button>
            </div>
          </form>
        </div>
      </SurveyContext.Provider>
    </>
  )
}

function Title() {
  const { survey, setSurvey } = useContext(SurveyContext);
  function handleChange(e) {
    setSurvey(prev => ({
      ...prev,
      title: e.target.value
    }))
  }
  return (
    <>
      <label>Title
        <input
        value={survey.title}
          onChange={handleChange}
        />
      </label>
    </>

  )
}

function SingleMutipleRanking({ index }) {
  const { survey, setSurvey } = useContext(SurveyContext);
  console.log(index)
  const titleQuestion = (e) => {
    setSurvey(prev => ({
      ...prev,
      questions: prev.questions.map((q, i) =>
        i === index
          ? { ...q, title: e.target.value }
          : q
      )
    }));
  };

  const addOption = (e) => {
    e.preventDefault();

    setSurvey(prev => ({
      ...prev,
      questions: prev.questions.map((q, i) =>
        i === index
          ? {
              ...q,
              options: {
                ...q.options,
                answers: [...q.options.answers, ""]
              }
            }
          : q
      )
    }));
  };

  const handleAddAnswers = (e, answerIndex) => {
    setSurvey(prev => ({
      ...prev,
      questions: prev.questions.map((q, i) =>
        i === index
          ? {
              ...q,
              options: {
                ...q.options,
                answers: q.options.answers.map((a, j) =>
                  j === answerIndex ? e.target.value : a
                )
              }
            }
          : q
      )
    }));
  };

  return (
    <>
      <label>
        TITLE question
        <input
          value={survey.questions[index].title}
          onChange={titleQuestion}
        />
      </label>

      <button type="button" onClick={addOption}>
        add Option
      </button>

      {survey.questions[index].options.answers.map((ans, i) => (
        <div className="options" key={i}>
          <input
            value={ans}
            onChange={(e) => handleAddAnswers(e, i)}
          />
        </div>
      ))}
    </>
  );
}

function Rating() {
  return (
    <>
      <h1>rating</h1>
    </>
  )
}

function Bool([index]) {
  const { survey, setSurvey } = useContext(SurveyContext);
  const titleQuestion = (e) =>{
    setSurvey( prev => ({
      ...prev,
      questions : prev.questions.map((q,i) => {
        i === index ? { ...q, title: e.target.value } : q
      })
    })   )
  }
  
  return (
    <>
      <label>
        TITLE question
        <input
          value={survey.questions[index].title}
          onChange={titleQuestion}
        />
      </label>
    </>
  )
}

function Open({index}) {
  const { survey, setSurvey } = useContext(SurveyContext);

  const titleQuestion = (e) => {
    setSurvey(prev => ({
      ...prev,
      questions: prev.questions.map((q, i) =>
        i === index
          ? { ...q, title: e.target.value }
          : q
      )
    }));
  };

  return (
    <>
      <label>
        TITLE question
        <input
          value={survey.questions[index].title}
          onChange={titleQuestion}
        />
      </label>
    </>
  )
}