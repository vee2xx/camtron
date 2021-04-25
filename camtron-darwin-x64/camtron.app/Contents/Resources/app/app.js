'use strict';
const electron = require('electron');
const dialog = electron.remote.dialog;
const menu = electron.remote.Menu;
const fs = require('fs');
const websocket = require('websocket-stream')
const isMac = process.platform === 'darwin'

let photoData;
let video;
let recordButton;
let isRecording = false;
var ws = null;
var connected = false;

let mediaRecorder; 
const recordedChunks = [];

var cameras = [];

function initialize () {
  video = window.document.querySelector('video');
  recordButton = window.document.querySelector('record');
  
  navigator.mediaDevices.enumerateDevices()
    .then(function(devices) {
      cameras = devices.filter(device => device.kind == 'videoinput')
      if (cameras.length > 0) {
        connectToCamera(cameras[0].deviceId);
      }
      addMenu();
    })
    .catch(function(err) {
      console.log(err.name +
         ": " + err.message);
    });
}

function addMenu() {

  var template = []

  var fileMenu = {label: 'File', submenu: [isMac ? { role: 'close' } : { role: 'quit' }]}

  template.push(fileMenu)

  var viewSubMenu = [{role: 'toggleDevTools'}]
  if (cameras.length > 0) {
    var camMenuItems = []
    cameras.forEach(function(cam) {
      camMenuItems.push({label: cam.label,
        click() { 
          connectToCamera(cam.deviceId) 
        } 
      })
    });
    
    viewSubMenu.push({label: 'Available webcanms', submenu: camMenuItems})
  }

  var viewMenu ={label: 'View', submenu: viewSubMenu}

  template.push(viewMenu)

  var camMenu = menu.buildFromTemplate(template); 
  menu.setApplicationMenu(camMenu); 
}

function connectToCamera(cameraId) {
  let errorCallback = (error) => {
    alert("Error accessing camera. Check to see if it is connected and access permissions have been granted")
  };
  window.navigator.webkitGetUserMedia({video: true, deviceId: { exact: cameraId }}, (localMediaStream) => {
    log("INFO", "Connecting to webcam...")
    video.src = window.URL.createObjectURL(localMediaStream);
    const options = { mimeType: 'video/webm; codecs=vp9' };
    mediaRecorder = new MediaRecorder(localMediaStream, options);

    mediaRecorder.ondataavailable = handleDataAvailable;
    mediaRecorder.onstop = handleStop;
    mediaRecorder.onerror = handleError;

    wsopen();
  }, errorCallback);
}

function shutdown() {
  mediaRecorder.stop()
  wsclose();
}

function wsopen() {
  ws = new WebSocket('ws://localhost:8080/streamVideo');
  ws.onmessage = onMessage;
  ws.onerror = onError;
  ws.onopen = onOpen;
  ws.onclose = onClose;
}

var onOpen = function(event) {
  connected = true;
  mediaRecorder.start(100);
}

var onClose = function(event) {
  connected = false;
}

var onMessage = function(event) {
  var data = event.data;
};

var onError = function(event) {
  log('ERROR', event.err )
};

function wsclose() {
  if (ws) {
      log('INFO', 'CLOSING ...');
      ws.close();
  }
  log('INFO', 'CLOSED: ');
  ws = null;
}

async function checkConnection(callback) {
  if (!connected) {
    wsopen();
    while (!connected) {
      await new Promise(r => setTimeout(r, 100));
    }
  }
  if (callback != null){
    callback();
  }  
}

function takePhotoAndSend() {
  let canvas = window.document.querySelector('canvas');
  canvas.getContext('2d').drawImage(video, 0, 0, 800, 600);
  photoData = canvas.toDataURL('image/png').replace(/^data:image\/(png|jpg|jpeg);base64,/, '');
  fetch("http://localhost:8080/upload", {
  headers: {'Content-Type': 'application/json'},
  body: JSON.stringify(photoData),
    "method": "POST"
  });
}

function handleDataAvailable(e) {
  checkConnection(function(){
    ws.send(e.data);
  });
}

function handleError(e) {
  log('ERROR', e.data)
}

async function handleStop(e) {
  fetch("http://localhost:8080/stop");
}

async function log(logLevel, message) {
  try {
    let logMessage = {logLevel: logLevel, message: message}
    fetch("http://localhost:8080/log", {
    headers: {'Content-Type': 'application/json'},
    body: JSON.stringify(logMessage),
      "method": "POST"
    });
  }
  catch(err) {
    alert("Unable to connect to server. " + err)
  }
}

window.onload = initialize;
window.onclose = shutdown;
