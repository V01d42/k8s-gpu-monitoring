import SearchIcon from '@mui/icons-material/Search';
import { alpha, AppBar, Box, InputBase, styled, Toolbar, Typography } from "@mui/material";

const SearchFieldWrapper = styled('div')(({ theme }) => ({
  position: 'relative',
  borderRadius: theme.shape.borderRadius,
  backgroundColor: alpha(theme.palette.common.white, 0.15),
  '&:hover': {
    backgroundColor: alpha(theme.palette.common.white, 0.25),
  },
  marginLeft: theme.spacing(1),
  width: '200px',
  [theme.breakpoints.up('sm')]: {
    marginLeft: 'auto',
    width: 'auto',
  },
}));

const SearchIconWrapper = styled('div')(({ theme }) => ({
  padding: theme.spacing(0, 2),
  height: '100%',
  position: 'absolute',
  pointerEvents: 'none',
  display: 'flex',
  alignItems: 'center',
  justifyContent: 'center',
}));

const StyledInputBase = styled(InputBase)(({ theme }) => ({
  color: 'inherit',
  width: '100%',
  '& .MuiInputBase-input': {
    padding: theme.spacing(1, 1, 1, 0),
    // vertical padding + font size from searchIcon
    paddingLeft: `calc(1em + ${theme.spacing(4)})`,
    transition: theme.transitions.create('width'),
    [theme.breakpoints.up('sm')]: {
      width: '25ch',
      '&:focus': {
        width: '30ch',
      },
    },
  },
}));

const TopMenu = () => {
    return (
        <>
        <Box sx={{ flexGrow: 1 }}>
        <AppBar position="fixed">
            <Toolbar>
                <Typography noWrap>GPU Monitoring</Typography>
                <SearchFieldWrapper>
                    <SearchIconWrapper>
                        <SearchIcon/>
                    </SearchIconWrapper>
                    <StyledInputBase placeholder="Hostname or GPU's name..."/>
                </SearchFieldWrapper>
            </Toolbar>
        </AppBar>
        </Box>
        </>
    );
}

export default TopMenu