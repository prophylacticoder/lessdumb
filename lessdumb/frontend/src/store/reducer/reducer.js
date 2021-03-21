import * as action_name from '../actions.js';

export function reducer(state = {ready: true, status: 'ready', sessionLength: 15, nBack: 1, btnPressed: [], results: {}}, action) {
  switch (action.type) {
    case action_name.CHANGE_READY:
      return {...state, ready: action.payload}
    case action_name.CHANGE_STATUS:
      switch (action.payload) {
        case 'playing':
          return {...state, status: 'playing'};
        case 'paused':
          return {...state, status: 'paused'};
        case 'ready':
          return {...state, status: 'ready'};
      }
      break;
    case action_name.CLEAR_CONTROL:
      return {...state, btnPressed: []};
    case action_name.CONTROL_PRESSED:
      return {...state, btnPressed: state.btnPressed.concat([action.payload])};
    case action_name.PAUSE:
      return {...state, status: 'paused'};
    case action_name.PLAYING:
      return {...state, status: 'playing'};
    case action_name.UPDATE:
      return {...state, sessionLength: action.payload.sessionLength ,nBack: action.payload.nBack, results: {...action.payload.results}}
    default:
      return state;
  }
}
