fetch('http://localhost:8081', {
  method: 'post',
  headers: {
    'Accept': 'application/json, text/plain, */*',
    'Content-Type': 'application/json'
  },
  body: JSON.stringify({user: "igor", op:"login", pwd:"xyz"})
}).then(res=>res.json())
  .then(res => console.log(res));
