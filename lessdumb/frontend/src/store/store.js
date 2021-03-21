import {createStore} from 'redux';
import {reducer} from './reducer/reducer.js'

export const store = createStore(reducer);
