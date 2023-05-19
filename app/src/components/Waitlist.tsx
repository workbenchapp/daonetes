import { Alert, Button, Form } from "react-bootstrap";
import { trackEvent, updateIntercom } from "next-intercom";
import { getCookie, setCookie } from 'cookies-next';
import { useState } from "react";

interface WaitlistProps {
    deviceKey: string;
}
export function Waitlist({ deviceKey }: WaitlistProps) {
    const waitlist = getCookie('waitlist', {});
    const waitlistEmail = getCookie('waitlist-email', {});
    const [email, setEmail] = useState(waitlistEmail?.toString() || "");

    if (waitlist && deviceKey && waitlist === deviceKey) {
        return (<Alert key="waitlist" variant="info">
            <h3>
                You're on our Waitlist
            </h3>
            Your device key is already in the waitlist - please hang on while we make sure its all working :)
        </Alert>)
    }
    if (waitlistEmail &&  waitlistEmail !== "") {
        return (<Alert key="waitlist" variant="info">
            <h3>
                You're on our Waitlist
            </h3>
            Your email ({email}) is already registered in our waitlist, but maybe you havn't Downloaded and installed the agent yet? - please hang on while we make sure its all working :)
        </Alert>)
    }
    return (
        <Alert key="waitlist" variant="info">
            <h3>
                Have you joined our Waitlist yet?
            </h3>
            <Form>
                <Form.Group className="mb-3" controlId="formBasicEmail">
                    <Form.Label>Email address</Form.Label>
                    <Form.Control
                        type="email"
                        value={email}
                        placeholder="Enter email"
                        onChange={(e) =>
                                        setEmail(e.target.value)
                                    }
                    />
                    <Form.Text className="text-muted">
                        We'll never share your email with anyone else.
                    </Form.Text>
                </Form.Group>

                <Form.Group
                    className="mb-3"
                    controlId="formBasicDeviceKey"
                    hidden
                >
                    <Form.Label>Device Key</Form.Label>
                    <Form.Control
                        type="string"
                        placeholder="detected DeviceKey"
                        value={deviceKey}
                        readOnly
                    />
                    <Form.Text className="text-muted">
                        DeviceKey for connecting to Vibenet.
                    </Form.Text>
                </Form.Group>


                <Button
                    variant="primary"
                    type="button"
                    onClick={() => {
                        updateIntercom("update", {
                            deviceKey: deviceKey,
                            email: email,
                        });
                        trackEvent("waitlist", {
                            deviceKey: deviceKey,
                            email: email,
                        })
                        setCookie('waitlist', deviceKey, {});
                        setCookie('waitlist-email', email, {});

                        // BOOM - need to refresh...
                    }
                    }
                >
                    Submit
                </Button>
            </Form>
        </Alert>
    )
}
