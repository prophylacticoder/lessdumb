import './App.scss';
import {reducer} from './store/reducer/reducer.js';
import * as actions from './store/actions.js';
import React from 'react';
import * as Redux from 'redux';
import * as ReactRedux from 'react-redux';
import {store} from './store/store.js';
import {letters} from './sounds/sounds.js';

const RR_Provider = ReactRedux.Provider;

class GameGrid extends React.Component {
  constructor(props) {
    super(props);

    this.state = {audio: {
      total: 0,
      correct: 0,
      incorrect: 0,
      match: false,
      session: []
      },
                 visual: {
      total: 0,
      correct: 0,
      incorrect: 0,
      match: false,
      session: []
      },
                  running: false,
                  counter: 0,
    }
   this.setSession = this.setSession.bind(this);
   this.playGame = this.playGame.bind(this);
  }

  componentDidMount() {
    for(let i = 0; i < letters.length; i++) {
      let audio = document.createElement('audio');
      audio.src = letters[i].link;
      audio.id = `sound${i}`;
      audio.preload = 'auto';
      document.getElementById('grid').appendChild(audio);
    }
    let audio = document.createElement('audio');
    audio.src = "https://dnbletters.s3-sa-east-1.amazonaws.com/levelup.wav";
    audio.id="level-up-audio";
    audio.preload = 'auto';
    document.getElementById('grid').appendChild(audio);
    this.setSession();
  }

  playGame() {
    switch (this.props.status) {
      case 'playing':
        this.props.changeStatus('paused');
        break;
      case 'paused':
        this.props.changeStatus('playing');
        break;
      case 'ready':
        let elem = document.getElementById('screen-phrase');
        elem.animate([
          {opacity: 1},
          {opacity: 0},


        ],
         {
          duration: 1500,
          easing: 'linear',
        }
          ).play();
        setTimeout(() => (elem.style.display = 'none'), 1400);
        this.props.changeStatus('playing');
        break;
    }
  }


  setSession() {
    let sessions = [[],[]];
    for (let i = 0, repeated = []; i < 2; i++, repeated = []) {
      // generate the random indexes
      for (let k = 0; k < 6; k++) {
        let randIndex;
        do
          randIndex = Math.floor(Math.random() * (this.props.sessionLength - this.props.nBack - 1)) + this.props.nBack + 1;
        while (repeated.includes(randIndex));
        repeated.push(randIndex);
      }

      for (let j = 0; j < this.props.sessionLength; j++) {
        // geerate the first nBack indexes
        if (this.props.nBack > j)
          sessions[i].push(Math.floor(Math.random() * 8));
        // created the patterns
        else if (repeated.includes(j))
          sessions[i].push(sessions[i][j - this.props.nBack]);
        // and finally randomize the rest of the array
        else {
          let rand;
          do
            rand = Math.floor(Math.random() * 8);
          while (sessions[i][j-this.props.nBack] == rand);
          sessions[i].push(rand);
        }


      }
    }
    this.setState({audio: {...this.state.audio, session: sessions[1]}, visual: {...this.state.visual, session: sessions[0]}});
  }


  render() {
    const loop = () => setTimeout(() => {
      if (this.props.status == 'playing') {
        const fBtnPressed = () => {
          if (this.props.btnPressed.length)
             if (this.props.btnPressed.indexOf('S') >= 0)
              if (this.state.visual.match) {
                this.setState({visual: {...this.state.visual, correct: this.state.visual.correct + 1, match: false}});
                  }
                  else this.setState({visual: {...this.state.visual, incorrect: this.state.visual.incorrect + 1}});
          if (this.props.btnPressed.indexOf('L') >= 0)
                  if (this.state.audio.match)
                    this.setState({audio: {...this.state.audio, correct: this.state.audio.correct + 1, match: false}});
                  else this.setState({audio: {...this.state.audio, incorrect: this.state.audio.incorrect + 1}});

          if (this.state.audio.match && this.props.btnPressed.indexOf('S') == -1)
            this.setState({audio: {...this.state.audio,incorrect: this.state.audio.incorrect + 1, match: false}});

          if (this.state.visual.match && this.props.btnPressed.indexOf('L') == -1)
            this.setState({visual: {...this.state.visual,incorrect: this.state.visual.incorrect + 1, match: false}});
        this.props.clearKey();
        }


        const transitionFunc = (id) => {
          const elem = document.getElementById(id);
          elem.animate(
          [
            {backgroundColor: '#0fbbd6e6'},
            {backgroundColor: '#800fd6'},
            {backgroundColor: '#0fbbd6e6'}

          ],
            {
             duration: 2500,
             easing: 'cubic-bezier(.23,.42,.29,.79)'
            }
          ).play();
        }
        if (this.state.counter < this.props.nBack) {
          transitionFunc(`cell${this.state.visual.session[this.state.counter]}`);
          document.getElementById(`sound${this.state.audio.session[this.state.counter]}`).play();
          fBtnPressed();
        }  else {
          fBtnPressed();
          if (this.state.audio.session[this.state.counter] == this.state.audio.session[this.state.counter - this.props.nBack]) {
            this.setState({audio: {...this.state.audio, match: true, total: this.state.audio.total + 1}});
            document.getElementById(`sound${this.state.audio.session[this.state.counter]}`).play();
          } else
            document.getElementById(`sound${this.state.audio.session[this.state.counter]}`).play();


          if (this.state.visual.session[this.state.counter] == this.state.visual.session[this.state.counter - this.props.nBack]) {
            transitionFunc(`cell${this.state.visual.session[this.state.counter]}`);
            this.setState({visual: {...this.state.visual, total: this.state.visual.total + 1, match: true}})
          } else
            transitionFunc(`cell${this.state.visual.session[this.state.counter]}`);
        }

        this.setState({counter: this.state.counter + 1});
        this.props.changeReady(true);

        if (this.state.counter < this.props.sessionLength)
          loop();
        else {
          setTimeout(() => {
            fBtnPressed();
            this.props.changeReady(true);
            let isLeveledUp = ((this.state.audio.correct + this.state.visual.correct) - (this.state.audio.incorrect + this.state.visual.incorrect)) / 12;

               isLeveledUp = (isLeveledUp >= 0.8) ? this.props.nBack + 1 : (isLeveledUp >= 0.25) ? this.props.nBack : (this.props.nBack == 1) ? 1 : this.props.nBack - 1;
            if (isLeveledUp > this.props.nBack) {
              document.getElementById('level-up-audio').play();

            }
            console.log({
              nBack: isLeveledUp,
              results: {
              audio: this.state.audio.total,
              visual: this.state.visual.total,
              audioCorrect: this.state.audio.correct,
              audioIncorrect: this.state.audio.incorrect,
              visualCorrect: this.state.visual.correct,
              visualIncorrect: this.state.visual.incorrect,
            },
             sessionLength: 10 + (isLeveledUp * 5),
          });
          this.props.update({
              nBack: isLeveledUp,
              results: {
                audio: this.state.audio.total,
                visual: this.state.visual.total,
                audioCorrect: this.state.audio.correct,
                audioIncorrect: this.state.audio.incorrect,
                visualCorrect: this.state.visual.correct,
                visualIncorrect: this.state.visual.incorrect,
            },
              sessionLength: 10 + (isLeveledUp * 5),
          })
          this.setState({
            running: false,
            counter: 0,
            audio: {
              correct: 0,
              incorrect: 0,
              total: 0
            },
            visual: {
              correct: 0,
              incorrect: 0,
              total: 0
            }
                        });
           this.setSession();
           this.props.changeStatus('ready');
          }, 3500);
        }
      } else if (this.state.running) this.setState({running: false});
    }, 3500);
    console.log('rendered!!');
    if (!this.state.running && this.props.status == 'playing') {
      console.log('rendered!!');
      loop();
      if (this.props.status == 'playing')
        this.setState({running: true});
    }
    return (
      <div id="gameMain">
        <h1>LEVEL = {this.props.nBack}</h1>
        <h2 id='screen-phrase'>Tap on the SCREEN or press SPACE-BAR to start a game!</h2>
        <div onClick={this.playGame} id="grid">
          <div id="cell0" />
          <div id="cell1" />
          <div id="cell2" />

          <div id="cell3" />
          <div id="center">
            <i id="playBtn" className="fas fa-play" />
          </div>
          <div id="cell4" />

          <div id="cell5" />
          <div id="cell6" />
          <div id="cell7" />
        </div>
      </div>
    )
  }
}

class GameControllers extends React.Component {
  constructor(props) {
    super(props);
    this.buttonPress = this.buttonPress.bind(this);
    this.state = {
      audioToggle: false,
      visualToggle: false
    }
  }

  componentDidMount() {
    document.addEventListener("keydown", this.buttonPress, false);
    let btnClick = document.createElement('audio');


    btnClick.src = 'https://dnbletters.s3-sa-east-1.amazonaws.com/click.wav';
    btnClick.preload = 'auto';
    btnClick.id = "btnClick";
    document.getElementById('controllers').appendChild(btnClick);


  }

  componentDidUpdate(nextProps, nextState) {
    if (nextProps.ready) {
      let visualElem = document.getElementById('visualBtn');
      let audioElem = document.getElementById('audioBtn');
        if (visualElem.classList.contains('active'))
           visualElem.classList.remove('active');
        if (audioElem.classList.contains('active'))
           audioElem.classList.remove('active');
    }
  }

  buttonPress(e) {
    let elem = document.getElementById('btnClick');
    // Get the portview's width
    const vw = Math.max(document.documentElement.clientWidth || 0, window.innerWidth || 0);
    elem.currentTime = 0;


    if (this.props.status == 'playing') {
      if ((e.target.id == 'visualBtn' || e.target.parentElement.id == 'visualBtn') && !this.state.visualToggle) {
        document.getElementById('visualBtn').classList.add('active');

        this.props.controlPressed('S');
        elem.play();
        this.setState({visualToggle: true});
      } else if ((e.target.id == 'audioBtn' || e.target.parentElement.id == 'audioBtn') && !this.state.audioToggle) {
        document.getElementById('audioBtn').classList.add('active');
        this.setState({audioToggle: true});
        elem.play();
        this.props.controlPressed('L');
      }

      switch (e.keyCode) {
        case 83:
          if (!this.state.visualToggle) {
            document.getElementById('visualBtn').classList.add('active')

            this.props.controlPressed('S');
            elem.play();
            this.setState({visualToggle: true});
          }
          break;
        case 76:
          if (!this.state.audioToggle) {
            document.getElementById('audioBtn').classList.add('active');
            this.setState({audioToggle: true});
            elem.play();
            this.props.controlPressed('L');
          }
          break;
        case 32:
          let playBtn = document.getElementById('playBtn');
          playBtn.classList.remove('fa-play');
          playBtn.classList.add('fa-pause');
          playBtn.style.fontSize = (vw > 700) ? '10em' : '5.5em';
          playBtn.style.opacity = '1';
          this.props.changeStatus('paused');
          break;
      }
    } else if (e.keyCode == 32) {
      let elem = document.getElementById('playBtn');
        switch (this.props.status) {
         case 'ready':
            elem.style.transition = 'font-size 1.5s, opacity, 1.5s';
            elem.style.fontSize = (vw > 700) ? '10em' : '5.5em';
            elem.style.opacity = '0';
            this.props.changeStatus('playing');
            break;
          case 'paused':
            elem.style.fontSize = (vw > 700) ? '10em' : '5.5em';
            elem.classList.remove('fa-pause');

            this.props.changeStatus('playing');
            break;
      }
    }
 }


  render() {
    if (this.props.ready) {
      this.props.changeReady(false);
      this.setState({audioToggle: false, visualToggle: false})
    }
    return (
      <div id="controllers">
        <button id="visualBtn" onClick={this.buttonPress}><i className="fas fa-eye" /></button>
        <button id="audioBtn" onClick={this.buttonPress}><i className="fas fa-headphones" /></button>
      </div>
    )
  }
}

const GameControllersConnected = ReactRedux.connect((state) => ({
  ready: state.ready,
  status: state.status,
}), (dispatch) => ({
  changeReady: (payload) => dispatch({type: actions.CHANGE_READY, payload}),
  changeStatus: payload => dispatch({type: actions.CHANGE_STATUS, payload}),
  controlPressed: payload => dispatch({type: actions.CONTROL_PRESSED, payload})
})
)(GameControllers);

const GameGridConnected = ReactRedux.connect((state) => ({
  btnPressed: state.btnPressed,
  nBack: state.nBack,
  ready: state.ready,
  sessionLength: state.sessionLength,
  status: state.status
}), (dispatch) => ({
  changeReady: (payload) => dispatch({type: actions.CHANGE_READY, payload}),
  changeStatus: payload => dispatch({type: actions.CHANGE_STATUS, payload}),
  clearKey: () => dispatch({type: actions.CLEAR_CONTROL}),
  update: payload => dispatch({type: actions.UPDATE, payload})
}))(GameGrid);

class App extends React.Component {
  constructor(props) {
    super(props);
  }

  render() {
    return (
      <RR_Provider store={store}>
        <GameGridConnected />
        <GameControllersConnected />
      </RR_Provider>
    )
  }
}

export default App;
