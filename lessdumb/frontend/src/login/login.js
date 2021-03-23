import React, {useState} from 'react';

// SCSS
import '../scss/login.scss'

const UseStateLogin = (props) => {
  const [logged, setLoginStatus] = useState(false);
  if (logged)
    console.log('do sth');
  else
  return (
    <div id="login-bar">
      <button id="loginBtn">LOGIN</button>
      <button id="registerBtn">REGISTER</button>
    </div>
  )
}

export default UseStateLogin;
