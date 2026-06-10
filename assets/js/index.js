import React from 'react'
import { createRoot } from 'react-dom/client'
import App from './Components/App'
import getInitialData from './getInitialData'
import Api from './Services/Api'

const root = createRoot(document.getElementById('root'))


getInitialData('api/settings').then((data) => {
  const api = new Api(data.settings);
  root.render(<App api={ api } { ...data } />);
})
