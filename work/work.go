package work

import (
	"net/http"

	"micro.mu"
)

var timer = `
<style>
  
.main { 
    padding-top: 100px;
    transition: transform 0.2s; 
    text-align: center; 
} 
  
.timer-circle { 
    border-radius: 50%; 
    width: 200px; 
    height: 200px; 
    margin: 20px auto; 
    display: flex; 
    align-items: center; 
    justify-content: center; 
    font-size: 25px; 
    color: #333333;
    border: 8px solid #333333;
} 
  
.control-buttons { 
    margin-top: 50px; 
    justify-content: space-evenly; 
} 
  
.control-buttons button { 
    background-color: darkslategrey;
    color: #fff; 
    border: none; 
    padding: 10px 20px; 
    border-radius: 5px; 
    cursor: pointer; 
} 
  
.control-buttons button:hover { 
    background-color: #333333; 
    transition: background-color 0.3s; 
}
</style>

<div class="main">
        <h1 id="title"></h1><br>
	<div class="timer-circle" 
             id="timer">60:00 
          </div> 
        <div class="control-buttons"> 
            <button onclick="togglePauseResume()"> 
                  Pause 
              </button> 
            <button onclick="restartTimer()"> 
                  Restart 
              </button> 
        </div> 
</div>

<script>
let timer; 
let minutes = 60; 
let seconds = 0; 
let isPaused = false; 
let enteredTime = null; 
  
function startTimer() { 
    // immediate update
    updateTimer()
    timer = setInterval(updateTimer, 1000); 
} 
  
function updateTimer() { 
    const timerElement = 
        document.getElementById('timer'); 
    timerElement.textContent =  
        formatTime(minutes, seconds); 
  
    if (minutes === 0 && seconds === 0) { 
        clearInterval(timer); 
        alert('Done'); 
    } else if (!isPaused) { 
        if (seconds > 0) { 
            seconds--; 
        } else { 
            seconds = 59; 
            minutes--; 
        } 
    } 
} 
  
function formatTime(minutes, seconds) { 
    return ` + "`${String(minutes).padStart(2, '0')}:${String(seconds).padStart(2, '0')}`" + `; 
} 
  
function togglePauseResume() { 
    const pauseResumeButton = 
        document.querySelector('.control-buttons button'); 
    isPaused = !isPaused; 
  
    if (isPaused) { 
        clearInterval(timer); 
        pauseResumeButton.textContent = 'Resume'; 
    } else { 
        startTimer(); 
        pauseResumeButton.textContent = 'Pause'; 
    } 
} 
  
function restartTimer() { 
    clearInterval(timer); 
    minutes = 60; 
    seconds = 0; 
    isPaused = false; 
    const timerElement = 
        document.getElementById('timer'); 
    timerElement.textContent = 
        formatTime(minutes, seconds); 
    const pauseResumeButton = 
        document.querySelector('.control-buttons button'); 
    pauseResumeButton.textContent = 'Pause'; 
    startTimer(); 
} 

function setTitle() {
	let title = prompt("Set your intention", "Work");
	if (title != null) {
		var t = document.getElementById("title");
		t.innerText = title;
	}
}
  
window.onload = (event) => { 
	setTitle()
	startTimer()
}
</script>
`

func Handler(w http.ResponseWriter, r *http.Request) {
	nav := `<a href="/work" class="head">New Job</a>`
	html := mu.Template("Work", "Do work by the hour", nav, timer)
	mu.Render(w, html)
}

func Register() {}
