from requests import post, Response, RequestException
import streamlit as st

SIGN_UP_REQUEST: str = "http://backend:8080/api/post/signup"

def sign_up():
    """
    Sends a signup request to the backend

    It will fail, if there's been an issue with the database or if the provided email is already in use.
    It will send JSON over http.
    The fields are defined as followed:

    ```
    {
        "name": str,
        "password": str,
        "email": str
    }
    ```

    The returning response will contain a statuscode and binary text.
    """

    st.set_page_config(
        page_title="Signup", 
        page_icon="./assets/Logo.png", 
        layout=None, 
        initial_sidebar_state=None, 
        menu_items=None
    )
    st.button("Back", on_click=lambda: st.session_state.update({"auth_page": "main"}), key="back_to_main")
    
    with st.form("Sign up"):
        username: str = st.text_input("Username")
        password: str = st.text_input("Password", type="password")
        email: str = st.text_input("Email")

        if st.form_submit_button("Sign up"):
            if not all([username, password, email]):
                st.warning("No field can be left empty")
            else:
                try:
                    response: Response = post(
                        url=SIGN_UP_REQUEST,
                        json={
                            "name": username,
                            "password": password,
                            "email": email
                        }
                    )
                    
                    if response.status_code != 200:
                        st.error(f"Sign up failed: {response.content.decode()}")
                    else:
                        st.success("Sign up succeeded. Please wait until an Admin has reviewed your request")
                        st.session_state.auth_page = 'main'
                        st.rerun()
                except RequestException as e:
                    st.error(f"Sign up failed: {str(e)}")