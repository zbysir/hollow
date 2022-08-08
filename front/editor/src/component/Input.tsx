import React, {ChangeEventHandler, HTMLAttributes, useEffect, useRef} from "react";

interface Props {
    type: string,
    placeholder?: string,
    label?: string,
    help?: string
    onChange?: ChangeEventHandler<HTMLInputElement>
    value: string,
    autoFocus?: boolean
    className?: string
}

const Index: React.FC<Props> = (props: Props & React.HTMLAttributes<Props>) => {
    const inputRef = useRef<HTMLInputElement>(null)

    useEffect(() => {
        if (props.autoFocus) {
            inputRef.current?.focus()
        }
    })
    return <div className="form-control w-full">
        {props.label ?
            <label className="label pl-0">
                <span className="label-text">{props.label}</span>
            </label>
            : null}
        <input
            ref={inputRef}
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
export default Index