const multiple = 10;
const mouseOverContainer = document.body;
const element = document.getElementById("element");

function transformElement(x, y) {
  let box = element.getBoundingClientRect();
  let calcX = -(y - box.y - (box.height / 2)) / multiple;
  let calcY = (x - box.x - (box.width / 2)) / multiple;
  console.log('yu',calcX)

  calcX = Math.min(Math.abs(calcX), 10) * ((calcX<0)?-1:1)
  calcY = Math.min(Math.abs(calcY), 10) * ((calcY<0)?-1:1)

  element.style.transform  = "rotateX("+ calcX +"deg) "
    + "rotateY("+ calcY +"deg)";
}

mouseOverContainer.addEventListener('mousemove', (e) => {
  window.requestAnimationFrame(function(){
    transformElement(e.clientX, e.clientY);
  });
});

mouseOverContainer.addEventListener('mouseleave', (e) => {
  window.requestAnimationFrame(function(){
    element.style.transform = "rotateX(0) rotateY(0)";
  });
});