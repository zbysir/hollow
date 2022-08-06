import {ChangeEventHandler} from "react";

interface Props {
    type: string,
    placeholder?: string,
    label?: string,
    help?: string
    onChange?: ChangeEventHandler<HTMLInputElement>
    value: string
}

export default function Input(props: Props) {
    return <div className="form-control w-full ">
        {props.label ?
            <label className="label">
                <span className="label-text">{props.label}</span>
            </label>
            : null}
        <input
            value={props.value}
            type={props.type}
            placeholder={props.placeholder}
            onInput={props.onChange}
            className="input input-bordered w-full input-sm "/>
        {props.help ?
            <label className="label">
                <span className="label-text-alt">{props.help}</span>
            </label>
            : null}
    </div>
}