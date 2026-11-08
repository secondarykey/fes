import cross from "./assets/crosshair.svg"
import back from "./assets/arrow-right-square-fill.svg"
import update from "./assets/arrow-clockwise.svg"
import terminal from "./assets/terminal.svg"
import fullscreen from "./assets/fullscreen.svg"
import fullscreenExit from "./assets/fullscreen-exit.svg"

const icons = {
  back: back,
  cross: cross,
  update: update,
  terminal: terminal,
  fullscreen: fullscreen,
  "fullscreen-exit": fullscreenExit,
};

//name でアイコンを選ぶ。name が変わればアイコンも変わる（全画面ボタンの切り替えなど）
const SVGButton = (props) => {

    const svg = icons[props.name];

    return (<>
      <button {...props} className={"svgBtn " + (props.className === undefined ? "" : props.className)}>
        {/* data を変えても読み直さないブラウザがあるので、アイコンごとに作り直す */}
        <object key={svg} type="image/svg+xml" data={svg} width={props.size} height={props.size}/>
      </button>
    </>)
}
export default SVGButton;
