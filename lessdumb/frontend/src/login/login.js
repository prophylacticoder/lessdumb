import React, {useState} from 'react';

import './scss/login.scss';

const handleSubmit = (e, username, password) => {
  e.preventDefault();
  console.log(e, username, password)
  const xhttp = new XMLHttpRequest();
  xhttp.open('POST', '', true);
  xhttp.send();
}

const UseStateLoginSection = () => {
  const {username, setUsername} = useState(null);
  const {password, setPassword} = useState(null);
  return (
    <div id="loginSection">
      <h1 style={{textAlign: 'center', marginTop: '20px'}}>LOGIN</h1>
      <form id="loginForm" onSubmit={handleSubmit}>
        <div>
          <label htmlFor="username">Username</label>
          <br/>
          <input type="text" id="loginUsername" name="username" value={username} onChange={setUsername}/>
        </div>
        <div>
          <label htmlFor="password">Password</label>
          <br/>
          <input type="password" id="loginPassword" name="password" value={password} onChange={setPassword}/>
        </div>
        <button type="submit">LOGIN</button>
      </form>
      <h1 style={{textAlign: 'center', marginTop: '20px'}}>lessDumb.org</h1>
    </div>
  )
}

export default UseStateLoginSection;
