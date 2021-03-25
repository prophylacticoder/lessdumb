import React, {useState} from 'react';
import LoginTab from './login';
import RegisterTab from './register'

// SCSS
import '../scss/login.scss';

const loginTabShow = () => {
  const registerTab = document.getElementById('registerSection');
  const elem = document.getElementById('loginSection');
  const registerDisplay = window.getComputedStyle(registerTab).getPropertyValue('display');
  const computedStyle = window.getComputedStyle(elem).getPropertyValue('display');

  if (registerDisplay === 'block')
    registerTab.style.display = 'none';

  if (computedStyle === 'block') {
    elem.style.display = 'none';
  } else {
    elem.style.display = 'block';
    document.getElementById('loginUsername').focus();
  }
}

const handleRegister = () => {
  const loginTab = document.getElementById('loginSection');
  const elem = document.getElementById('registerSection');
  const loginDisplay = window.getComputedStyle(loginTab).getPropertyValue('display');
  const computedStyle = window.getComputedStyle(elem).getPropertyValue('display');

  if (loginDisplay === 'block')
      loginTab.style.display = 'none';

  if (computedStyle === 'block') {
    elem.style.display = 'none';
  } else {
    elem.style.display = 'block';
    document.getElementById('registerUsername').focus();
  }
}

const UseStateLogin = (props) => {
  const [logged, setLoginStatus] = useState(false);
  if (logged)
    console.log('do sth');
  else
  return (
    <>
      <LoginTab />
      <RegisterTab />
      <div id="login-bar">
        <button id="loginBtn" onClick={loginTabShow}>LOGIN</button>
        <button id="registerBtn" onClick={handleRegister}>REGISTER</button>
      </div>
    </>
  )
}

export default UseStateLogin;
